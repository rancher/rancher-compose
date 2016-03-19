package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/rancher/rancher-catalog-service/manager"
	"github.com/rancher/rancher-catalog-service/service"
	"net/http"
)

func main() {
	log.Infof("Starting Rancher Catalog service")
	router := service.NewRouter()
	handler := service.MuxWrapper{false, router}

	go func() {
		manager.SetEnv()
		manager.Init()
	}()
	log.Fatal(http.ListenAndServe(":8088", &handler))
}
