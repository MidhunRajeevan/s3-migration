package config

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type s3ArchiveConfig struct {
	Location      string
	Endpoint      string
	Bucket        string
	AccessKey     string
	SecretKey     string
	UseSSL        bool
	AllowInsecure bool
}

// S3Archive Configuration
var S3Archive s3ArchiveConfig

const (
	s3ArchiveLocation      = "S3_ARCHIVE_LOCATION"
	s3ArchiveEndpoint      = "S3_ARCHIVE_ENDPOINT"
	s3ArchiveBucket        = "S3_ARCHIVE_BUCKET"
	s3ArchiveAccessKey     = "S3_ARCHIVE_ACCESS_KEY"
	s3ArchiveSecretKey     = "S3_ARCHIVE_SECRET_KEY"
	s3ArchiveUseSSL        = "S3_ARCHIVE_USE_SSL"
	s3ArchiveAllowInsecure = "S3_ARCHIVE_ALLOW_INSECURE"
)

// S3ArchiveClient for uploads
var S3ArchiveClient *minio.Client

// InitializeS3Archive connection
func InitializeS3Archive() {
	var err error
	var ok bool

	S3Archive.Location, ok = os.LookupEnv(s3ArchiveEndpoint)
	if !ok {
		S3Archive.Location = "us-east-1"
	}

	S3Archive.Endpoint, ok = os.LookupEnv(s3ArchiveEndpoint)
	if !ok {
		panic("S3_ARCHIVE_ENDPOINT environment variable required but not set")
	}

	S3Archive.Bucket, ok = os.LookupEnv(s3ArchiveBucket)
	if !ok {
		panic("S3_ARCHIVE_BUCKET environment variable required but not set")
	}

	S3Archive.AccessKey, ok = os.LookupEnv(s3ArchiveAccessKey)
	if !ok {
		panic("S3_ARCHIVE_ACCESS_KEY environment variable required but not set")
	}

	S3Archive.SecretKey, ok = os.LookupEnv(s3ArchiveSecretKey)
	if !ok {
		panic("S3_ARCHIVE_SECRET_KEY environment variable required but not set")
	}

	S3Archive.UseSSL = true
	useSSL, ok := os.LookupEnv(s3ArchiveUseSSL)
	if ok && (strings.ToLower(useSSL) == "false") {
		S3Archive.UseSSL = false
	}

	S3Archive.AllowInsecure = false
	allowInsecure, ok := os.LookupEnv(s3ArchiveAllowInsecure)
	if ok && (strings.ToLower(allowInsecure) == "true") {
		S3Archive.AllowInsecure = true
	}

	transport := http.DefaultTransport
	if S3Archive.AllowInsecure {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Initialize minio client object.
	log.Println("Connecting to S3Archive")
	S3ArchiveClient, err = minio.New(S3Archive.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(S3Archive.AccessKey, S3Archive.SecretKey, ""),
		Secure:    S3Archive.UseSSL,
		Transport: transport,
	})
	if err != nil {
		log.Println("S3Archive Client Error")
		log.Printf("%v\n", S3Archive)
		log.Fatalln(err)
	}

	// Check bucket
	log.Println("Checking Bucket Exists")
	found, err := S3ArchiveClient.BucketExists(context.Background(), S3Archive.Bucket)
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		log.Println("S3Archive Bucket Error")
		log.Println(errResponse)
		log.Fatalln(err)
	} else if !found {
		log.Fatalln("Bucket does not exist!")
	}

	log.Println("S3Archive Configuration Complete")
}
