package jobs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gramlane/internal/agent"
	"gramlane/internal/chain"
	"gramlane/internal/framing"
	"gramlane/internal/httpcache"
	"gramlane/internal/names"
	"gramlane/internal/post"
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
	Job        Job         `json:"job"`
	Quote      quote.Quote `json:"quote"`
	Paid       string      `json:"paid"`
	Payer      string      `json:"payer,omitempty"`
	Wallet     string      `json:"wallet,omitempty"`
	OnChain    bool        `json:"onChain"`
	Explorer   string      `json:"explorer,omitempty"`
	TxNote     string      `json:"txNote,omitempty"`
	Settlement string      `json:"settlement,omitempty"`
	Remaining  uint64      `json:"gramsRemaining,omitempty"`
	Output     string      `json:"output"`
	Note       string      `json:"note"`
}

var Catalog = []Job{
	{ID: "resolve", Name: "Resolve a Kasdomain", Blurb: "Look up a .kas name. The gram bill is the sequenced call, not the name.", Grams: 12_000, Lane: "SIGN1", Kind: "resolve"},
	{ID: "rank", Name: "L1 balance depth", Blurb: "Owner from indexer, then api.kaspa.org balance. Still grams, still L1.", Grams: 18_000, Lane: "SIGN1", Kind: "rank"},
	{ID: "dag", Name: "BlockDAG heartbeat", Blurb: "Read virtual DAA from api.kaspa.org. Cheap inclusion probe.", Grams: 5_000, Lane: "SEQ1", Kind: "dag"},
	{ID: "profile", Name: "Pull Kasdomain profile texts", Blurb: "Avatar, x, website if the public name index has them.", Grams: 22_000, Lane: "SIGN1", Kind: "profile"},
	{ID: "batch", Name: "Batch three resolves", Blurb: "kaspadao.kas + gramlane.kas + kachat.kas. One voucher burn.", Grams: 40_000, Lane: "SIGN1", Kind: "batch"},
	{ID: "vault", Name: "Vault bump (not the lock)", Blurb: "Grams pay the framing action. The vault still locks KAS. Worked #234: amount 1 can read as 264.", Grams: 8_000, Lane: "SEQ1", Kind: "vault"},
	{ID: "postage", Name: "KaChat postage", Blurb: "Sequenced stamp to a .kas contact. Not E2E — KaChat seals ciph_msg in the wallet.", Grams: 9_000, Lane: "MSG1", Kind: "postage"},
	{ID: "agent", Name: "AI agent call", Blurb: "HTTP 402 for a machine. Grok if XAI_API_KEY is set; otherwise local tools. Grams pay the call.", Grams: 25_000, Lane: "AGENT", Kind: "agent"},
	{ID: "site", Name: "Kasdomain page", Blurb: "Hang a living webpage on a name. Prefer /kasdomain to publish headline and about.", Grams: 15_000, Lane: "SIGN1", Kind: "site"},
}

type Fit struct {
	Job  Job    `json:"job"`
	Runs uint64 `json:"runs"`
}

func Fits(grams uint64) []Fit {
	var out []Fit
	for _, j := range Catalog {
		if j.Grams > 0 && grams >= j.Grams {
			out = append(out, Fit{Job: j, Runs: grams / j.Grams})
		}
	}
	return out
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
	return httpcache.Get(url, func() ([]byte, error) {
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
	})
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
		name = "kaspadao"
	}
	return "https://api.knsdomains.org/mainnet/api/v1/" + name + ".kas/owner"
}

func profileURL(name string) string {
	name = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(name)), ".kas")
	if name == "" {
		name = "kaspadao"
	}
	return "https://api.knsdomains.org/mainnet/api/v1/domain/" + name + ".kas/profile"
}

func prepaidToken(paid string) bool {
	switch strings.ToLower(strings.TrimSpace(paid)) {
	case "grams", "prepaid", "work-credit", "kaspa-work-credit":
		return true
	}
	return false
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
		Note:   "HTTP receipt only. This dApp does not verify a WorkCredit UTXO spend on L1.",
	}
	switch {
	case prepaidToken(paid):
		r.Settlement = "prepaid-grams"
		r.TxNote = "prepaid sequencer grams — not a new KAS send"
		r.Note = "HTTP receipt only. This dApp does not verify a WorkCredit UTXO spend on L1. Local bar tab until consume()."
	case chain.IsTxID(paid):
		look := chain.Find(paid)
		r.Explorer = look.Explorer
		r.OnChain = look.Found
		r.Settlement = "kas-fallback"
		if look.Found {
			r.TxNote = "api.kaspa.org has this tx. Open Explorer."
		} else if look.Err != "" {
			r.TxNote = "txid shape, not in api yet (index lag) or " + look.Err
		}
	default:
		r.Settlement = "http-receipt"
		r.TxNote = "not a txid — HTTP receipt only, not on the explorer"
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
		for _, n := range []string{"kaspadao.kas", "gramlane.kas", "kachat.kas"} {
			b, err := getJSON(ownerURL(n))
			if err != nil {
				parts = append(parts, n+": "+err.Error())
				continue
			}
			parts = append(parts, n+":\n"+pretty(b))
		}
		r.Output = strings.Join(parts, "\n\n")
	case "vault":
		v := framing.Demo()
		if hx := strings.TrimSpace(q); hx != "" && !strings.HasSuffix(strings.ToLower(hx), ".kas") {
			got, err := framing.DecodeHex(hx)
			if err != nil {
				return r, err
			}
			v.Custom = &got
		}
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return r, err
		}
		r.Output = string(b)
	case "postage":
		st := post.StampMsg(r.Payer, q, j.Grams)
		b, err := json.MarshalIndent(st, "", "  ")
		if err != nil {
			return r, err
		}
		r.Output = string(b)
	case "agent":
		name, prompt := splitAgentQ(q)
		if prompt == "" {
			c, err := agent.CardFor(name)
			if err != nil {
				return r, err
			}
			b, err := json.MarshalIndent(c, "", "  ")
			if err != nil {
				return r, err
			}
			r.Output = string(b)
			break
		}
		rep, err := agent.Ask(prompt, name)
		if err != nil {
			return r, err
		}
		b, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			return r, err
		}
		r.Output = string(b)
	case "site":
		p, err := names.Lookup(q)
		if err != nil {
			return r, err
		}
		b, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			return r, err
		}
		r.Output = string(b)
	default:
		return r, fmt.Errorf("unknown job")
	}
	return r, nil
}

func splitAgentQ(q string) (name, prompt string) {
	q = strings.TrimSpace(q)
	if q == "" {
		return "kaspadao.kas", "What is Gramlane?"
	}
	if i := strings.Index(q, " | "); i >= 0 {
		return strings.TrimSpace(q[:i]), strings.TrimSpace(q[i+3:])
	}
	low := strings.ToLower(q)
	if strings.HasSuffix(low, ".kas") && !strings.Contains(q, " ") {
		return names.Normalize(q), ""
	}
	return "", q
}
