package config

import (
	"os"
	"strconv"
	"strings"
)

type appConfig struct {
	TenantString  string
	UploadLimit   int64
	ContentTypes  []string
	ListenPort    int
	AllowInsecure bool
}

// App configuration from environment
var App appConfig

const (
	appListenPort    = "APP_LISTEN_PORT"
	appTenantString  = "APP_TENANT_STRING"
	appUploadLimit   = "APP_UPLOAD_LIMIT"
	appAllowInsecure = "APP_ALLOW_INSECURE"
)

const (
	defaultListenPort   = 9090
	defaultTenantString = "tenants"
	defaultUploadLimit  = 10
)

// InitializeApp Configuration
func InitializeApp() {
	var err error

	// Listen Port
	if it, ok := os.LookupEnv(appListenPort); ok {
		if App.ListenPort, err = strconv.Atoi(it); err != nil {
			App.ListenPort = defaultListenPort
		}
	} else {
		App.ListenPort = defaultListenPort
	}

	// Upload Limit
	if it, ok := os.LookupEnv(appTenantString); ok {
		App.TenantString = it
	} else {
		App.TenantString = defaultTenantString
	}

	// Upload Limit
	if it, ok := os.LookupEnv(appUploadLimit); ok {
		if App.UploadLimit, err = strconv.ParseInt(it, 10, 64); err != nil {
			App.UploadLimit = defaultUploadLimit
		}
	} else {
		App.UploadLimit = defaultUploadLimit
	}

	App.ContentTypes = []string{"image/jpeg", "image/jpg", "image/png", "image/gif", "application/pdf", "application/octet-stream"}

	// Allow Insecure
	App.AllowInsecure = false
	insecure, ok := os.LookupEnv(appAllowInsecure)
	if ok && (strings.ToLower(insecure) == "true") {
		App.AllowInsecure = true
	}

}
