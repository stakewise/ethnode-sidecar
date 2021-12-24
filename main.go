package main

import (
	"github.com/gorilla/mux"
	"github.com/stakewise/operator-sidecar/clients"
	"github.com/stakewise/operator-sidecar/config"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	prysm := clients.NewEth2Client("prysm")
	lighthouse := clients.NewEth2Client("lighthouse")
	teku := clients.NewEth2Client("teku")
	nimbus := clients.NewEth2Client("nimbus")
	geth := clients.NewEth1Client("geth")
	erigon := clients.NewEth1Client("erigon")
	router := mux.NewRouter()
	router.HandleFunc("/prysm/readiness", prysm.Readiness).Methods(http.MethodGet)
	router.HandleFunc("/prysm/liveness", prysm.Liveness).Methods(http.MethodGet)
	router.HandleFunc("/lighthouse/readiness", lighthouse.Readiness).Methods(http.MethodGet)
	router.HandleFunc("/lighthouse/liveness", lighthouse.Liveness).Methods(http.MethodGet)
	router.HandleFunc("/teku/readiness", teku.Readiness).Methods(http.MethodGet)
	router.HandleFunc("/teku/liveness", teku.Liveness).Methods(http.MethodGet)
	router.HandleFunc("/nimbus/readiness", nimbus.Readiness).Methods(http.MethodGet)
	router.HandleFunc("/nimbus/liveness", nimbus.Liveness).Methods(http.MethodGet)
	router.HandleFunc("/geth/readiness", geth.HealthCheck).Methods(http.MethodGet)
	router.HandleFunc("/geth/liveness", geth.HealthCheck).Methods(http.MethodGet)
	router.HandleFunc("/erigon/readiness", erigon.HealthCheck).Methods(http.MethodGet)
	router.HandleFunc("/erigon/liveness", erigon.HealthCheck).Methods(http.MethodGet)

	srv := &http.Server{
		Handler:      router,
		Addr:         cfg.Server.BindAddr,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
