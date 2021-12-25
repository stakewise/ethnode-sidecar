package clients

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/stakewise/ethnode-sidecar/common/hexutil"
	"github.com/stakewise/ethnode-sidecar/config"
	"log"
	"net/http"
)

type eth1Client struct {
	cfg    *config.Config
	addr   string
	client *resty.Client
}

func NewEth1Client() *eth1Client {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	var addr string
	addr = cfg.Client.Scheme + "://" + cfg.Client.Host + ":" + cfg.Client.Port

	client := resty.New()
	return &eth1Client{
		cfg:    cfg,
		addr:   addr,
		client: client,
	}
}

// HealthCheck Returns OK if the node is fully
// synchronized and ready to receive traffic
func (e *eth1Client) HealthCheck(w http.ResponseWriter, r *http.Request) {
	var ethSyncing = struct {
		Jsonrpc string      `json:"jsonrpc"`
		ID      int         `json:"id"`
		Result  interface{} `json:"result"`
	}{}

	_, err := e.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}`).
		SetResult(&ethSyncing).
		Post(e.addr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: "+err.Error())
	}

	switch ethSyncing.Result.(type) {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	case bool:
		ethSyncingResult := ethSyncing.Result
		if ethSyncingResult == false {
			fmt.Fprintf(w, "StatusOK. ETH1 node is healthy.")
		}
	case map[string]interface{}:
		ethSyncingResult := ethSyncing.Result.(map[string]interface{})
		if hBlock, ok := ethSyncingResult["highestBlock"]; ok {
			highestBlock, err := hexutil.DecodeUint64(fmt.Sprintf("%s", hBlock))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			currentBlock, err := hexutil.DecodeUint64(fmt.Sprintf("%s", ethSyncingResult["currentBlock"]))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			if highestBlock-currentBlock > 50 {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				fmt.Fprintf(w, "StatusOK. ETH1 node is healthy.")
			}
		} else {
			fmt.Fprintf(w, "StatusOK. ETH1 node is healthy.")
		}
	}

}
