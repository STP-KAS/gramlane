// Package names pulls a .kas identity from the live KNS indexer.
// Uniqueness is indexer FCFS, not consensus.
package names

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const indexer = "https://api.knsdomains.org/mainnet/api/v1"

type Page struct {
	Name     string `json:"name"`
	Owner    string `json:"owner,omitempty"`
	AssetID  string `json:"assetId,omitempty"`
	PayURI   string `json:"payUri,omitempty"`
	KasURI   string `json:"kasUri,omitempty"`
	DID      string `json:"did,omitempty"`
	Bio      string `json:"bio,omitempty"`
	Avatar   string `json:"avatarUrl,omitempty"`
	Banner   string `json:"banner,omitempty"`
	Website  string `json:"website,omitempty"`
	Redirect string `json:"redirectUrl,omitempty"`
	X        string `json:"x,omitempty"`
	GitHub   string `json:"github,omitempty"`
	Telegram string `json:"telegram,omitempty"`
	Email    string `json:"email,omitempty"`
	Warning  string `json:"warning"`
}

func Normalize(raw string) string {
	n := strings.ToLower(strings.TrimSpace(raw))
	n = strings.TrimPrefix(n, "kas://")
	n = strings.TrimPrefix(n, "did:kas:")
	n = strings.TrimSuffix(n, "/")
	if n == "" {
		return "kns.kas"
	}
	if !strings.Contains(n, ".") {
		n += ".kas"
	}
	return n
}

func Lookup(raw string) (*Page, error) {
	name := Normalize(raw)
	p := &Page{
		Name:    name,
		KasURI:  "kas://" + name,
		DID:     "did:kas:" + name,
		Warning: "Indexer-backed. Not consensus uniqueness. This is a generated site from profile texts, not IPFS.",
	}
	own, err := getJSON(indexer + "/" + url.PathEscape(name) + "/owner")
	if err != nil {
		return p, err
	}
	var env struct {
		Success bool `json:"success"`
		Data    struct {
			ID      string `json:"id"`
			AssetID string `json:"assetId"`
			Owner   string `json:"owner"`
		} `json:"data"`
	}
	if json.Unmarshal(own, &env) != nil || env.Data.Owner == "" {
		return p, fmt.Errorf("no owner for %s", name)
	}
	p.Owner = env.Data.Owner
	p.AssetID = env.Data.AssetID
	if p.AssetID == "" {
		p.AssetID = env.Data.ID
	}
	p.PayURI = "kaspa:" + strings.TrimPrefix(p.Owner, "kaspa:")
	if p.AssetID != "" {
		keys := "redirectUrl,bio,avatarUrl,website,x,github,telegram,discord,email,banner"
		path := indexer + "/domain/" + url.PathEscape(p.AssetID) + "/profile?keys=" + url.QueryEscape(keys)
		if b, err := getJSON(path); err == nil {
			fillProfile(p, b)
		}
	}
	return p, nil
}

func fillProfile(p *Page, b []byte) {
	var env struct {
		Success bool `json:"success"`
		Data    struct {
			Profile map[string]any `json:"profile"`
		} `json:"data"`
	}
	if json.Unmarshal(b, &env) != nil {
		return
	}
	g := func(k string) string {
		v, ok := env.Data.Profile[k]
		if !ok || v == nil {
			return ""
		}
		return strings.TrimSpace(fmt.Sprint(v))
	}
	p.Bio = g("bio")
	p.Avatar = httpURL(g("avatarUrl"))
	p.Banner = httpURL(g("banner"))
	p.Website = httpURL(g("website"))
	p.Redirect = httpURL(g("redirectUrl"))
	p.Email = g("email")
	p.X = xURL(g("x"))
	p.GitHub = githubURL(g("github"))
	tg := g("telegram")
	if tg != "" {
		tg = strings.TrimPrefix(tg, "@")
		if strings.HasPrefix(tg, "http") {
			p.Telegram = httpURL(tg)
		} else {
			p.Telegram = "https://t.me/" + tg
		}
	}
}

func httpURL(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://") {
		return s
	}
	return ""
}

func xURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	s = strings.TrimPrefix(s, "@")
	s = strings.TrimPrefix(s, "x.com/")
	s = strings.TrimPrefix(s, "twitter.com/")
	return "https://x.com/" + s
}

func githubURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	s = strings.TrimPrefix(s, "@")
	return "https://github.com/" + s
}

var httpc = &http.Client{Timeout: 12 * time.Second}

func getJSON(u string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "gramlane/1.0")
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
		return nil, fmt.Errorf("indexer %s", res.Status)
	}
	return b, nil
}
