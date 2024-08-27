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

type targets3Config struct {
	Location      string
	Endpoint      string
	Bucket        string
	AccessKey     string
	SecretKey     string
	UseSSL        bool
	AllowInsecure bool
}

// S3 Configuration
var Target targets3Config

const (
	s3Location      = "S3_TARGET_LOCATION"
	s3Endpoint      = "S3_TARGET_ENDPOINT"
	s3Bucket        = "S3_TARGET_BUCKET"
	s3AccessKey     = "S3_TARGET_ACCESS_KEY"
	s3SecretKey     = "S3_TARGET_SECRET_KEY"
	s3UseSSL        = "S3_TARGET_USE_SSL"
	s3AllowInsecure = "S3_TARGET_ALLOW_INSECURE"
)

// S3Client for uploads
var TargetClient *minio.Client

// InitializeS3 connection
func InitializeTarget() {
	var err error
	var ok bool

	Target.Location, ok = os.LookupEnv(s3Endpoint)
	if !ok {
		Target.Location = "us-east-1"
	}

	Target.Endpoint, ok = os.LookupEnv(s3Endpoint)
	if !ok {
		panic("Target_ENDPOINT environment variable required but not set")
	}

	Target.Bucket, ok = os.LookupEnv(s3Bucket)
	if !ok {
		panic("S3_TARGET_BUCKET environment variable required but not set")
	}

	Target.AccessKey, ok = os.LookupEnv(s3AccessKey)
	if !ok {
		panic("S3_TARGET_ACCESS_KEY environment variable required but not set")
	}

	Target.SecretKey, ok = os.LookupEnv(s3SecretKey)
	if !ok {
		panic("S3_TARGET_SECRET_KEY environment variable required but not set")
	}

	Target.UseSSL = true
	useSSL, ok := os.LookupEnv(s3UseSSL)
	if ok && (strings.ToLower(useSSL) == "false") {
		Target.UseSSL = false
	}

	Target.AllowInsecure = false
	allowInsecure, ok := os.LookupEnv(s3AllowInsecure)
	if ok && (strings.ToLower(allowInsecure) == "true") {
		Target.AllowInsecure = true
	}

	transport := http.DefaultTransport
	if Target.AllowInsecure {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Initialize minio client object.
	log.Println("Connecting to S3")
	TargetClient, err = minio.New(Target.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(Target.AccessKey, Target.SecretKey, ""),
		Secure:    Target.UseSSL,
		Transport: transport,
	})
	if err != nil {
		log.Println("S3 Client Error")
		log.Printf("%v\n", Target)
		log.Fatalln(err)
	}

	// Check bucket
	log.Println("Checking Bucket Exists")
	found, err := TargetClient.BucketExists(context.Background(), Target.Bucket)
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
