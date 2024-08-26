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

type s3Config struct {
	Location      string
	Endpoint      string
	Bucket        string
	AccessKey     string
	SecretKey     string
	UseSSL        bool
	AllowInsecure bool
}

// S3 Configuration
var S3 s3Config

const (
	s3Location      = "S3_LOCATION"
	s3Endpoint      = "S3_ENDPOINT"
	s3Bucket        = "S3_BUCKET"
	s3AccessKey     = "S3_ACCESS_KEY"
	s3SecretKey     = "S3_SECRET_KEY"
	s3UseSSL        = "S3_USE_SSL"
	s3AllowInsecure = "S3_ALLOW_INSECURE"
)

// S3Client for uploads
var S3Client *minio.Client

// InitializeS3 connection
func InitializeS3() {
	var err error
	var ok bool

	S3.Location, ok = os.LookupEnv(s3Endpoint)
	if !ok {
		S3.Location = "us-east-1"
	}

	S3.Endpoint, ok = os.LookupEnv(s3Endpoint)
	if !ok {
		panic("S3_ENDPOINT environment variable required but not set")
	}

	S3.Bucket, ok = os.LookupEnv(s3Bucket)
	if !ok {
		panic("S3_BUCKET environment variable required but not set")
	}

	S3.AccessKey, ok = os.LookupEnv(s3AccessKey)
	if !ok {
		panic("S3_ACCESS_KEY environment variable required but not set")
	}

	S3.SecretKey, ok = os.LookupEnv(s3SecretKey)
	if !ok {
		panic("S3_SECRET_KEY environment variable required but not set")
	}

	S3.UseSSL = true
	useSSL, ok := os.LookupEnv(s3UseSSL)
	if ok && (strings.ToLower(useSSL) == "false") {
		S3.UseSSL = false
	}

	S3.AllowInsecure = false
	allowInsecure, ok := os.LookupEnv(s3AllowInsecure)
	if ok && (strings.ToLower(allowInsecure) == "true") {
		S3.AllowInsecure = true
	}

	transport := http.DefaultTransport
	if S3.AllowInsecure {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Initialize minio client object.
	log.Println("Connecting to S3")
	S3Client, err = minio.New(S3.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(S3.AccessKey, S3.SecretKey, ""),
		Secure:    S3.UseSSL,
		Transport: transport,
	})
	if err != nil {
		log.Println("S3 Client Error")
		log.Printf("%v\n", S3)
		log.Fatalln(err)
	}


	// Check bucket
	log.Println("Checking Bucket Exists")
	found, err := S3Client.BucketExists(context.Background(), S3.Bucket)
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		log.Println("S3 Bucket Error")
		log.Println(errResponse)
		log.Fatalln(err)
	} else if !found {
		log.Fatalln("Bucket does not exist!")
	}

	log.Println("S3 Configuration  Complete")

}
