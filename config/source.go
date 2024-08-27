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

type sourceS3Config struct {
	Location      string
	Endpoint      string
	Bucket        string
	AccessKey     string
	SecretKey     string
	UseSSL        bool
	AllowInsecure bool
}

var Source sourceS3Config

const (
	sourceLocation      = "S3_SOURCE_LOCATION"
	sourceEndpoint      = "S3_SOURCE_ENDPOINT"
	sourceBucket        = "S3_SOURCE_BUCKET"
	sourceAccessKey     = "S3_SOURCE_ACCESS_KEY"
	sourceSecretKey     = "S3_SOURCE_SECRET_KEY"
	sourceUseSSL        = "S3_SOURCE_USE_SSL"
	sourceAllowInsecure = "S3_SOURCE_ALLOW_INSECURE"
)

// sourceClient for uploads
var SourceClient *minio.Client

// Initializesource connection
func Initializesource() {
	var err error
	var ok bool

	Source.Location, ok = os.LookupEnv(sourceEndpoint)
	if !ok {
		Source.Location = "us-east-1"
	}

	Source.Endpoint, ok = os.LookupEnv(sourceEndpoint)
	if !ok {
		panic("S3_ARCHIVE_ENDPOINT environment variable required but not set")
	}

	Source.Bucket, ok = os.LookupEnv(sourceBucket)
	if !ok {
		panic("S3_ARCHIVE_BUCKET environment variable required but not set")
	}

	Source.AccessKey, ok = os.LookupEnv(sourceAccessKey)
	if !ok {
		panic("S3_ARCHIVE_ACCESS_KEY environment variable required but not set")
	}

	Source.SecretKey, ok = os.LookupEnv(sourceSecretKey)
	if !ok {
		panic("S3_ARCHIVE_SECRET_KEY environment variable required but not set")
	}

	Source.UseSSL = true
	useSSL, ok := os.LookupEnv(sourceUseSSL)
	if ok && (strings.ToLower(useSSL) == "false") {
		Source.UseSSL = false
	}

	Source.AllowInsecure = false
	allowInsecure, ok := os.LookupEnv(sourceAllowInsecure)
	if ok && (strings.ToLower(allowInsecure) == "true") {
		Source.AllowInsecure = true
	}

	transport := http.DefaultTransport
	if Source.AllowInsecure {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Initialize minio client object.
	log.Println("Connecting to source")
	SourceClient, err = minio.New(Source.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(Source.AccessKey, Source.SecretKey, ""),
		Secure:    Source.UseSSL,
		Transport: transport,
	})
	if err != nil {
		log.Println("source Client Error")
		log.Printf("%v\n", Source)
		log.Fatalln(err)
	}

	// Check bucket
	log.Println("Checking Bucket Exists")
	found, err := SourceClient.BucketExists(context.Background(), Source.Bucket)
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		log.Println("source Bucket Error")
		log.Println(errResponse)
		log.Fatalln(err)
	} else if !found {
		log.Fatalln("Bucket does not exist!")
	}

	log.Println("source Configuration Complete")
}
