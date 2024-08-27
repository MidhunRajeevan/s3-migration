package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	app "github.com/MidhunRajeevan/s3-migration/app"
	config "github.com/MidhunRajeevan/s3-migration/config"
)

func main() {
	config.Initializesource()
	config.InitializeTarget()
	config.InitializeApp()
	config.InitializeDB()

	if config.App.AllowInsecure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	http.HandleFunc(fmt.Sprintf("/%s", config.App.TenantString), app.Uploads)
	http.HandleFunc(fmt.Sprintf("/%s/", config.App.TenantString), app.Uploads)
	http.HandleFunc("/", app.Index)

	http.HandleFunc("/start", app.StartMigrationHandler)
	http.HandleFunc("/stop", app.StopMigrationHandler)

	url := fmt.Sprintf(":%d", config.App.ListenPort)
	log.Println("Starting server at " + url)
	log.Fatal(http.ListenAndServe(url, nil))
}
