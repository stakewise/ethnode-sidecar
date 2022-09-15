package clients

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/stakewise/ethnode-sidecar/common/hexutil"
	"github.com/stakewise/ethnode-sidecar/config"
)

type eth1Client struct {
	cfg               *config.Config
	addr              string
	client            *resty.Client
	authorizationType AuthorizationMethod
	jwtSecret         string
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
		cfg:               cfg,
		addr:              addr,
		client:            client,
		authorizationType: AuthorizationMethod(cfg.Client.AuthorizationType),
		jwtSecret:         cfg.Client.JWTSecret,
	}
}

// HealthCheck Returns OK if the node is fully
// synchronized and ready to receive traffic
func (e *eth1Client) HealthCheck(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Jsonrpc string      `json:"jsonrpc"`
		ID      int         `json:"id"`
		Result  interface{} `json:"result"`
	}

	var ethSyncing, ethPeersConnected response
	authorizationHeaders := map[string]string{}

	if e.authorizationType == Bearer {
		token, err := CreateJWTAuthToken(e.jwtSecret)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		authorizationHeaders["Authorization"] = fmt.Sprintf("Bearer %s", token)
	}

	_, err := e.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeaders(authorizationHeaders).
		SetBody(`{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}`).
		SetResult(&ethSyncing).
		Post(e.addr)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = e.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeaders(authorizationHeaders).
		SetBody(`{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":74}`).
		SetResult(&ethPeersConnected).
		Post(e.addr)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Error: "+err.Error())
		return
	}

	peers, err := hexutil.DecodeUint64(fmt.Sprintf("%s", ethPeersConnected.Result))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if peers < 3 {
		fmt.Println("Number of connected peers less than 3...NODE NOT READY")
		w.WriteHeader(http.StatusInternalServerError)
		return
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
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			currentBlock, err := hexutil.DecodeUint64(fmt.Sprintf("%s", ethSyncingResult["currentBlock"]))
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if highestBlock-currentBlock > 50 {
				fmt.Println(fmt.Sprintf("highestBlock-currentBlock < 50, highestBlock: %d, currentBlock: %d", highestBlock, currentBlock))
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				fmt.Fprintf(w, "StatusOK. ETH1 node is healthy.")
			}
		} else {
			fmt.Fprintf(w, "StatusOK. ETH1 node is healthy.")
		}
	}

}
