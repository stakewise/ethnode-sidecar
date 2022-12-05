package clients

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/stakewise/ethnode-sidecar/config"
)

type eth2Client struct {
	cfg    *config.Config
	addr   string
	client *resty.Client
}

func NewEth2Client() *eth2Client {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	var addr string
	addr = cfg.Client.Scheme + "://" + cfg.Client.Host + ":" + cfg.Client.Port

	client := resty.New()
	return &eth2Client{
		cfg:    cfg,
		addr:   addr,
		client: client,
	}
}

// Readiness Returns OK if the node is fully
// synchronized and ready to receive traffic
func (e *eth2Client) Readiness(w http.ResponseWriter, r *http.Request) {
	var data = struct {
		Data struct {
			HeadSlot     string `json:"head_slot"`
			SyncDistance string `json:"sync_distance"`
			IsSyncing    bool   `json:"is_syncing"`
		} `json:"data"`
	}{}
	var dataError = struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{}

	_, err := e.client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&data).
		SetError(&dataError).
		Get(e.addr + "/eth/v1/node/syncing")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if data.Data.IsSyncing {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		fmt.Fprintf(w, "StatusOK. Beacon node is synced.")
	}
}

// Liveness Returns OK if the node is healthy,
// synchronized and ready to receive traffic
func (e *eth2Client) Liveness(w http.ResponseWriter, r *http.Request) {
	var data = struct {
		Data struct {
			HeadSlot     string `json:"head_slot"`
			SyncDistance string `json:"sync_distance"`
			IsSyncing    bool   `json:"is_syncing"`
		} `json:"data"`
	}{}
	var dataError = struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{}
	var dataCheckpointSync = struct {
		Data struct {
			BackFillSyncing struct {
				Completed int `json:"completed"`
				Remaining int `json:"remaining"`
			} `json:"BackFillSyncing"`
		} `json:"data"`
	}{}

	_, err := e.client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&data).
		SetError(&dataError).
		Get(e.addr + "/eth/v1/node/syncing")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, err = e.client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&dataCheckpointSync).
		SetError(&dataError).
		Get(e.addr + "/eth/v1/node/syncing")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	resp, err := e.client.R().
		SetHeader("Content-Type", "application/json").
		Get(e.addr + "/eth/v1/node/health")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if dataCheckpointSync.Data.BackFillSyncing.Remaining != 0 {
		if dataCheckpointSync.Data.BackFillSyncing.Remaining-dataCheckpointSync.Data.BackFillSyncing.Completed < 50 {
			fmt.Fprintf(w, "StatusOK. Beacon node is healthy.")
		} else {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		syncDistance, _ := strconv.Atoi(data.Data.SyncDistance)
		if resp.StatusCode() == http.StatusOK && syncDistance < 50 {
			fmt.Fprintf(w, "StatusOK. Beacon node is healthy.")
		} else {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
