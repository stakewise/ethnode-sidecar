package main

import (
	"github.com/gorilla/mux"
	"github.com/stakewise/ethnode-sidecar/clients"
	"github.com/stakewise/ethnode-sidecar/config"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	eth1 := clients.NewEth1Client()
	eth2 := clients.NewEth2Client()
	router.HandleFunc("/eth1/readiness", eth1.HealthCheck).Methods(http.MethodGet)
	router.HandleFunc("/eth1/liveness", eth1.HealthCheck).Methods(http.MethodGet)
	router.HandleFunc("/eth2/readiness", eth2.Readiness).Methods(http.MethodGet)
	router.HandleFunc("/eth2/liveness", eth2.Liveness).Methods(http.MethodGet)

	srv := &http.Server{
		Handler:      router,
		Addr:         cfg.Server.BindAddr,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
