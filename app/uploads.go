package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/MidhunRajeevan/s3-migration/config"
	"github.com/MidhunRajeevan/s3-migration/util"
	"github.com/h2non/filetype"
	"github.com/minio/minio-go/v7"
)

func getUploadDetails(w http.ResponseWriter, r *http.Request) {
	var err error

	type metadata struct {
		Hash string `json:"hash"`
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	ctx := context.Background()
	s3Client := config.SourceClient

	segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	re, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Println("name_process_error", err)
		util.InternalServerError(&w, "name_process_error")
		return
	}
	objDir := re.ReplaceAllString(strings.ToLower(segments[1]), "-")

	objName := fmt.Sprintf("%s/%s", objDir, segments[3])

	objInfo, err := s3Client.StatObject(ctx, config.Source.Bucket, objName, minio.StatObjectOptions{})
	if err != nil {
		log.Println("get_object_error", err)
		util.NotFound(&w, "get_object_error")
		return
	}

	keys := []string{"etag", "name", "lastModified", "size", "contentType", "userMetadata"}
	var objStat map[string]interface{}
	objStr, _ := json.Marshal(objInfo)
	json.Unmarshal(objStr, &objStat)
	var objMeta metadata
	for key, val := range objStat {
		if !util.Contains(keys, key) {
			delete(objStat, key)
		} else if key == "userMetadata" {
			objMetaStr, _ := json.Marshal(val)
			json.Unmarshal(objMetaStr, &objMeta)
			delete(objStat, "userMetadata")
		}
	}
	objStat["metadata"] = objMeta

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(objStat)
}

func getUploads(w http.ResponseWriter, r *http.Request) {
	var err error

	ctx := context.Background()
	s3Client := config.SourceClient

	segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	re, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Println("name_process_error", err)
		util.InternalServerError(&w, "name_process_error")
		return
	}
	objDir := re.ReplaceAllString(strings.ToLower(segments[1]), "-")

	objName := fmt.Sprintf("%s/%s", objDir, segments[3])

	var objInfo minio.ObjectInfo
	object, err := s3Client.GetObject(ctx, config.Source.Bucket, objName, minio.GetObjectOptions{})
	if err != nil {
		log.Println("get_object_error", err)
		util.NotFound(&w, "get_object_error")
		return
	}
	defer object.Close()

	objInfo, err = object.Stat()
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", objInfo.ContentType)
		if _, err = io.Copy(w, object); err != nil {
			log.Println("object_copy_error", err)
			util.InternalServerError(&w, "object_copy_error")
			return
		}
		return
	}
}

func postUploads(w http.ResponseWriter, r *http.Request) {
	var err error

	ctx := context.Background()
	s3Client := config.TargetClient

	segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	re, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Println("name_process_error", err)
		util.InternalServerError(&w, "name_process_error")
		return
	}
	objDir := re.ReplaceAllString(strings.ToLower(segments[1]), "-")

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Println("form_parse_error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// for every file in multipart
	var response []map[string]string
	for _, fhs := range r.MultipartForm.File {
		for _, fh := range fhs {
			var f multipart.File
			if f, err = fh.Open(); nil != err {
				log.Println("file_open_error", err)
				util.BadRequest(&w, "file_open_error")
				return
			}

			content, err := ioutil.ReadAll(f)
			if err != nil {
				log.Println("file_read_error", err)
				util.BadRequest(&w, "file_read_error")
				return
			}

			hasher := sha256.New()
			if _, err := io.Copy(hasher, bytes.NewReader(content)); err != nil {
				log.Println("hash_compute_error", err)
				util.InternalServerError(&w, "hash_compute_error")
				return
			}
			objHash := fmt.Sprintf("%x", hasher.Sum(nil))
			objExt := filepath.Ext(fh.Filename)
			objName := fmt.Sprintf("%s/%s%s", objDir, objHash, objExt)
			objFile := fmt.Sprintf("%s%s", objHash, objExt)
			objURL := fmt.Sprintf("/%s/%s/%s/%s", segments[0], segments[1], segments[2], objFile)
			objSize := fh.Size

			// validate
			if objSize > config.App.UploadLimit {
				log.Println("file_too_big", objSize)
				util.BadRequest(&w, "file_too_big")
				return
			}

			contentType := fh.Header.Get("Content-Type")
			if len(contentType) > 0 && !util.Contains(config.App.ContentTypes, contentType) {
				log.Println("content_not_acceptable", contentType)
				util.BadRequest(&w, "content_not_acceptable")
				return
			}

			kind, _ := filetype.Match(content)
			if len(kind.MIME.Value) <= 0 && !util.Contains(config.App.ContentTypes, kind.MIME.Value) {
				log.Println("content_not_acceptable", kind.MIME.Value)
				util.BadRequest(&w, "content_not_acceptable")
				return
			}

			userMetadata := make(map[string]string)
			userMetadata["name"] = url.QueryEscape(fh.Filename)
			userMetadata["hash"] = objHash
			userMetadata["url"] = objURL

			opts := minio.PutObjectOptions{ContentType: contentType, UserMetadata: userMetadata}
			_, err = s3Client.PutObject(ctx, config.Target.Bucket, objName, bytes.NewReader(content), objSize, opts)
			if err != nil {
				log.Println("s3_put_error", err)
				util.InternalServerError(&w, "s3_put_error")
				return
			}

			response = append(response, userMetadata)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Uploads API
func Uploads(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.Path)

	segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if !((segments[0] == config.App.TenantString) && (segments[2] == "uploads")) {
		util.NotFound(&w, "path_not_found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		switch len(segments) {
		case 4: // /tenants/1/uploads/1
			getUploads(w, r)
		case 5: // /tenants/1/uploads/1/details
			getUploadDetails(w, r)
		default:
			util.NotFound(&w, "path_not_found")
			return
		}
	case http.MethodPost:
		switch len(segments) {
		case 3: // /tenants/1/uploads
			postUploads(w, r)
		default:
			util.NotFound(&w, "path_not_found")
			return
		}
	default:
		util.MethodNotAllowed(&w, "method_not_allowed")
	}
}
