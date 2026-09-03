package jobs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gramlane/internal/quote"
)

type Job struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Blurb string `json:"blurb"`
	Grams uint64 `json:"grams"`
	Lane  string `json:"lane"`
	Kind  string `json:"kind"`
}

type Receipt struct {
	Job    Job         `json:"job"`
	Quote  quote.Quote `json:"quote"`
	Paid   string      `json:"paid"`
	Payer  string      `json:"payer,omitempty"`
	Wallet string      `json:"wallet,omitempty"`
	Output string      `json:"output"`
	Note   string      `json:"note"`
}

var Catalog = []Job{
	{ID: "resolve", Name: "Resolve a .kas name", Blurb: "Live KNS indexer lookup. The gram bill is the sequenced call, not the name.", Grams: 12_000, Lane: "KNS1", Kind: "resolve"},
	{ID: "rank", Name: "L1 balance depth", Blurb: "Owner from indexer, then api.kaspa.org balance. Still grams, still L1.", Grams: 18_000, Lane: "KNS1", Kind: "rank"},
	{ID: "dag", Name: "BlockDAG heartbeat", Blurb: "Read virtual DAA from api.kaspa.org. Cheap inclusion probe.", Grams: 5_000, Lane: "SEQ1", Kind: "dag"},
	{ID: "profile", Name: "Pull KNS profile texts", Blurb: "Avatar, x, website if the indexer has them.", Grams: 22_000, Lane: "KNS1", Kind: "profile"},
	{ID: "batch", Name: "Batch three resolves", Blurb: "kns.kas + kaspa.kas + kachat.kas. One voucher burn.", Grams: 40_000, Lane: "KNS1", Kind: "batch"},
}

func Get(id string) (Job, bool) {
	for _, j := range Catalog {
		if j.ID == id {
			return j, true
		}
	}
	return Job{}, false
}

func QuoteJob(j Job) (quote.Quote, error) {
	return quote.Grams(j.Grams, j.Lane)
}

var httpc = &http.Client{Timeout: 12 * time.Second}

func getJSON(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	res, err := httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("%s: %s", res.Status, truncate(string(b), 180))
	}
	return b, nil
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func pretty(b []byte) string {
	var v any
	if json.Unmarshal(b, &v) != nil {
		return truncate(string(b), 800)
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return truncate(string(b), 800)
	}
	return string(out)
}

func ownerURL(name string) string {
	name = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(name)), ".kas")
	if name == "" {
		name = "kns"
	}
	return "https://api.knsdomains.org/mainnet/api/v1/" + name + ".kas/owner"
}

func profileURL(name string) string {
	name = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(name)), ".kas")
	if name == "" {
		name = "kns"
	}
	return "https://api.knsdomains.org/mainnet/api/v1/domain/" + name + ".kas/profile"
}

func Run(j Job, q string, paid string) (Receipt, error) {
	return RunAs(j, q, paid, "", "")
}

func RunAs(j Job, q, paid, payer, wallet string) (Receipt, error) {
	qq, err := QuoteJob(j)
	if err != nil {
		return Receipt{}, err
	}
	r := Receipt{
		Job:    j,
		Quote:  qq,
		Paid:   paid,
		Payer:  strings.TrimSpace(payer),
		Wallet: strings.TrimSpace(wallet),
		Note:   "HTTP receipt only. This dApp does not verify a WorkCredit UTXO spend on L1. Payer is the connected wallet address if you sent one.",
	}
	switch j.Kind {
	case "resolve":
		b, err := getJSON(ownerURL(q))
		if err != nil {
			return r, err
		}
		r.Output = pretty(b)
	case "profile":
		b, err := getJSON(profileURL(q))
		if err != nil {
			return r, err
		}
		r.Output = pretty(b)
	case "rank":
		b, err := getJSON(ownerURL(q))
		if err != nil {
			return r, err
		}
		var env struct {
			Success bool `json:"success"`
			Data    struct {
				Owner string `json:"owner"`
			} `json:"data"`
		}
		if err := json.Unmarshal(b, &env); err != nil {
			return r, err
		}
		owner := env.Data.Owner
		if owner == "" {
			r.Output = pretty(b)
			return r, fmt.Errorf("no owner")
		}
		bal, err := getJSON("https://api.kaspa.org/addresses/" + owner + "/balance")
		if err != nil {
			return r, err
		}
		r.Output = "owner:\n" + pretty(b) + "\n\nbalance:\n" + pretty(bal)
	case "dag":
		b, err := getJSON("https://api.kaspa.org/info/blockdag")
		if err != nil {
			return r, err
		}
		r.Output = pretty(b)
	case "batch":
		var parts []string
		for _, n := range []string{"kns.kas", "kaspa.kas", "kachat.kas"} {
			b, err := getJSON(ownerURL(n))
			if err != nil {
				parts = append(parts, n+": "+err.Error())
				continue
			}
			parts = append(parts, n+":\n"+pretty(b))
		}
		r.Output = strings.Join(parts, "\n\n")
	default:
		return r, fmt.Errorf("unknown job")
	}
	return r, nil
}
