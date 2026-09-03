package chain

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type UTXO struct {
	Address string `json:"address"`
	TxID    string `json:"txid"`
	Index   int    `json:"index"`
	Amount  uint64 `json:"amount"`
}

func AddressUTXOs(addr string) ([]UTXO, error) {
	u := "https://api.kaspa.org/addresses/" + url.PathEscape(addr) + "/utxos"
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "gramlane/1.0")
	cli := &http.Client{Timeout: 12 * time.Second}
	res, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("api.kaspa.org %s", res.Status)
	}
	var raw []struct {
		Address  string `json:"address"`
		Outpoint struct {
			TransactionID string `json:"transactionId"`
			Index         int    `json:"index"`
		} `json:"outpoint"`
		Entry struct {
			Amount string `json:"amount"`
		} `json:"utxoEntry"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	out := make([]UTXO, 0, len(raw))
	for _, r := range raw {
		var n uint64
		fmt.Sscanf(r.Entry.Amount, "%d", &n)
		out = append(out, UTXO{Address: r.Address, TxID: r.Outpoint.TransactionID, Index: r.Outpoint.Index, Amount: n})
	}
	return out, nil
}
