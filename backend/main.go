package main

import (
	"example.com/chemical-plant-permit-service/config"
	"example.com/chemical-plant-permit-service/health"
	"example.com/chemical-plant-permit-service/httpapi"
	"example.com/chemical-plant-permit-service/store"
	"example.com/chemical-plant-permit-service/web"
	"log"
	"net/http"
)

func main() {
	c := config.Load()
	m := http.NewServeMux()
	m.HandleFunc("/healthz", health.Handler)
	m.Handle("/api/v1/", httpapi.New(store.New()))
	m.HandleFunc("/", web.Handler)
	log.Printf("chemical permit service listening on %s", c.Address())
	log.Fatal(serveAddress(c.Address(), m))
}
