package names

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const linksPath = "data/id-links.json"

type Identity struct {
	Address string   `json:"address"`
	Names   []string `json:"names"`
	Linked  string   `json:"linked,omitempty"`
	Display string   `json:"display"`
	Total   int      `json:"total"`
	Note    string   `json:"note"`
}

var (
	linkMu sync.Mutex
	links  map[string]string
)

func ForAddress(addr string) (*Identity, error) {
	addr = strings.TrimSpace(addr)
	if !strings.HasPrefix(addr, "kaspa:") {
		return nil, fmt.Errorf("kaspa address")
	}
	id := &Identity{
		Address: addr,
		Display: addr,
		Note:    "Indexer-backed. Linking one .kas name here is Gramlane’s choice for this address, not a consensus primary.",
	}
	q := url.Values{}
	q.Set("owner", addr)
	q.Set("type", "domain")
	q.Set("pageSize", "20")
	for page := 1; page <= 3; page++ {
		q.Set("page", fmt.Sprintf("%d", page))
		b, err := getJSON(indexer + "/assets?" + q.Encode())
		if err != nil {
			if page == 1 {
				return id, err
			}
			break
		}
		var env struct {
			Success bool `json:"success"`
			Data    struct {
				Assets []struct {
					Asset string `json:"asset"`
					Owner string `json:"owner"`
				} `json:"assets"`
				Pagination struct {
					TotalItems int `json:"totalItems"`
					TotalPages int `json:"totalPages"`
				} `json:"pagination"`
			} `json:"data"`
		}
		if json.Unmarshal(b, &env) != nil {
			break
		}
		id.Total = env.Data.Pagination.TotalItems
		for _, a := range env.Data.Assets {
			n := strings.ToLower(strings.TrimSpace(a.Asset))
			if n == "" {
				continue
			}
			id.Names = append(id.Names, n)
		}
		if page >= env.Data.Pagination.TotalPages || env.Data.Pagination.TotalPages == 0 {
			break
		}
	}
	id.Linked = Linked(addr)
	if id.Linked != "" {
		id.Display = id.Linked
	} else if len(id.Names) == 1 {
		id.Display = id.Names[0]
	}
	return id, nil
}

func Linked(addr string) string {
	loadLinks()
	linkMu.Lock()
	defer linkMu.Unlock()
	return links[strings.ToLower(strings.TrimSpace(addr))]
}

func Link(addr, name string) error {
	addr = strings.TrimSpace(addr)
	name = strings.ToLower(strings.TrimSpace(name))
	if !strings.HasPrefix(addr, "kaspa:") {
		return fmt.Errorf("kaspa address")
	}
	if name == "" {
		loadLinks()
		linkMu.Lock()
		delete(links, strings.ToLower(addr))
		err := saveLinksLocked()
		linkMu.Unlock()
		return err
	}
	name = Normalize(name)
	id, err := ForAddress(addr)
	if err != nil {
		return err
	}
	ok := false
	for _, n := range id.Names {
		if n == name {
			ok = true
			break
		}
	}
	if !ok {
		p, lerr := Lookup(name)
		if lerr != nil || !strings.EqualFold(p.Owner, addr) {
			return fmt.Errorf("%s is not owned by this kaspa address", name)
		}
	}
	loadLinks()
	linkMu.Lock()
	defer linkMu.Unlock()
	if links == nil {
		links = map[string]string{}
	}
	links[strings.ToLower(addr)] = name
	return saveLinksLocked()
}

func loadLinks() {
	linkMu.Lock()
	defer linkMu.Unlock()
	if links != nil {
		return
	}
	links = map[string]string{}
	b, err := os.ReadFile(linksPath)
	if err != nil {
		return
	}
	_ = json.Unmarshal(b, &links)
}

func saveLinksLocked() error {
	if err := os.MkdirAll(filepath.Dir(linksPath), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(links, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(linksPath, b, 0o644)
}
