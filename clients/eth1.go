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

func NewEth1Client(ethClient string) *eth1Client {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	var addr string
	switch ethClient {
	case "geth":
		addr = cfg.Client.Geth.Scheme + "://" + cfg.Client.Geth.Host + ":" + cfg.Client.Geth.Port
	case "erigon":
		addr = cfg.Client.Erigon.Scheme + "://" + cfg.Client.Erigon.Host + ":" + cfg.Client.Erigon.Port
	}

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
	var currentBlock = struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  string `json:"result"`
	}{}

	var highestBlock = struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  struct {
			BaseFeePerGas    string        `json:"baseFeePerGas"`
			Difficulty       string        `json:"difficulty"`
			ExtraData        string        `json:"extraData"`
			GasLimit         string        `json:"gasLimit"`
			GasUsed          string        `json:"gasUsed"`
			Hash             string        `json:"hash"`
			LogsBloom        string        `json:"logsBloom"`
			Miner            string        `json:"miner"`
			MixHash          string        `json:"mixHash"`
			Nonce            string        `json:"nonce"`
			Number           string        `json:"number"`
			ParentHash       string        `json:"parentHash"`
			ReceiptsRoot     string        `json:"receiptsRoot"`
			Sha3Uncles       string        `json:"sha3Uncles"`
			Size             string        `json:"size"`
			StateRoot        string        `json:"stateRoot"`
			Timestamp        string        `json:"timestamp"`
			TotalDifficulty  string        `json:"totalDifficulty"`
			Transactions     []string      `json:"transactions"`
			TransactionsRoot string        `json:"transactionsRoot"`
			Uncles           []interface{} `json:"uncles"`
		} `json:"result"`
	}{}

	_, err := e.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`).
		SetResult(&currentBlock).
		Post(e.addr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: "+err.Error())
	}

	_, err = e.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}`).
		SetResult(&highestBlock).
		Post(e.addr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	currentBlockDecoded, err := hexutil.DecodeUint64(currentBlock.Result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	highestBlockDecoded, err := hexutil.DecodeUint64(highestBlock.Result.Number)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if highestBlockDecoded-currentBlockDecoded > 50 || highestBlockDecoded == 0 {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		fmt.Fprintf(w, "%s %s", currentBlockDecoded, highestBlockDecoded)
		fmt.Fprintf(w, "StatusOK. ETH1 node is healthy.")
	}
}
