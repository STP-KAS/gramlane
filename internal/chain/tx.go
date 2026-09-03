// Package chain looks up a Kaspa mainnet tx so a receipt can say on-chain or not.
package chain

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const Explorer = "https://explorer.kaspa.org/txs/"
const apiTx = "https://api.kaspa.org/transactions/"

var txidRe = regexp.MustCompile(`(?i)^[0-9a-f]{64}$`)

func IsTxID(s string) bool {
	return txidRe.MatchString(strings.TrimSpace(s))
}

type Lookup struct {
	TxID     string `json:"txid"`
	Explorer string `json:"explorer"`
	Found    bool   `json:"found"`
	Block    string `json:"blockHash,omitempty"`
	Err      string `json:"error,omitempty"`
}

func Find(txid string) Lookup {
	txid = strings.TrimSpace(txid)
	out := Lookup{TxID: txid, Explorer: Explorer + txid}
	if !IsTxID(txid) {
		out.Err = "not a 64-hex txid"
		return out
	}
	req, err := http.NewRequest(http.MethodGet, apiTx+txid, nil)
	if err != nil {
		out.Err = err.Error()
		return out
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "gramlane/1.0")
	cli := &http.Client{Timeout: 12 * time.Second}
	res, err := cli.Do(req)
	if err != nil {
		out.Err = err.Error()
		return out
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 400 {
		out.Err = fmt.Sprintf("api.kaspa.org %s", res.Status)
		return out
	}
	var v map[string]any
	if json.Unmarshal(b, &v) != nil {
		out.Found = true
		return out
	}
	out.Found = true
	if h, ok := v["block_hash"].(string); ok {
		out.Block = h
	}
	if h, ok := v["accepting_block_hash"].(string); ok && out.Block == "" {
		out.Block = h
	}
	return out
}
