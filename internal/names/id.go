package names

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gramlane/internal/appenv"
)

var linksPath = appenv.File("id-links.json")

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
		Note:    "One Kasdomain per Kaspa address on Gramlane. A name is a shop sign. The address is the door. Live names come from the public name index.",
	}
	q := url.Values{}
	q.Set("owner", addr)
	q.Set("type", "domain")
	q.Set("pageSize", "100")
	for page := 1; page <= 5; page++ {
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
	id.Names = sortNames(id.Names)
	id.Linked = Linked(addr)
	if id.Linked != "" {
		id.Display = id.Linked
	} else if len(id.Names) == 1 {
		id.Display = id.Names[0]
	}
	return id, nil
}

func sortNames(in []string) []string {
	if len(in) < 2 {
		return in
	}
	score := func(n string) int {
		switch {
		case n == "kdao.kas" || n == "kaspadao.kas":
			return 0
		case strings.HasPrefix(n, "kdao"):
			return 1
		case strings.Contains(n, "kdao") || strings.Contains(n, "kaspadao"):
			return 2
		default:
			return 3
		}
	}
	out := append([]string(nil), in...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if score(out[j]) < score(out[i]) || (score(out[j]) == score(out[i]) && out[j] < out[i]) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// PickOwned turns "kdao" into an owned name. Exact match first, then unique substring.
func PickOwned(addr, q string) (string, []string, error) {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return "", nil, fmt.Errorf("name")
	}
	want := Normalize(q)
	id, err := ForAddress(addr)
	if err != nil {
		return "", nil, err
	}
	needle := strings.TrimSuffix(q, ".kas")
	var hits []string
	for _, n := range id.Names {
		if n == want {
			return n, nil, nil
		}
		if knsMatch(n, needle) {
			hits = append(hits, n)
		}
	}
	if p, lerr := Lookup(want); lerr == nil && strings.EqualFold(p.Owner, addr) {
		return want, nil, nil
	}
	if len(hits) == 1 {
		return hits[0], nil, nil
	}
	if len(hits) > 1 {
		return "", hits, fmt.Errorf("several names match %s — pick one", q)
	}
	return "", nil, fmt.Errorf("%s is not owned by this kaspa address", want)
}

func knsMatch(name, needle string) bool {
	if needle == "" {
		return false
	}
	if strings.Contains(name, needle) {
		return true
	}
	if needle == "kdao" && (strings.Contains(name, "kaspadao") || strings.HasPrefix(name, "kdao")) {
		return true
	}
	return false
}

func Linked(addr string) string {
	loadLinks()
	linkMu.Lock()
	defer linkMu.Unlock()
	return links[strings.ToLower(strings.TrimSpace(addr))]
}

// Face is this wallet's kasdomain identity. Not KNS. Names come from the held book.
func Face(addr string) *Identity {
	addr = strings.TrimSpace(addr)
	id := &Identity{
		Address: addr,
		Names:   []string{},
		Display: addr,
		Note:    "kasdomain face. First registered is the default. Other names on this wallet are custody. Same wallet, switch shop.",
	}
	if !strings.HasPrefix(addr, "kaspa:") {
		id.Note = "kaspa address"
		return id
	}
	b := Book(addr)
	if b == nil {
		id.Note = "No kasdomain face yet. Fund a name. The kaspa address stays the door."
		return id
	}
	for _, h := range b.Names {
		id.Names = append(id.Names, h.Name)
	}
	id.Total = len(id.Names)
	if b.Primary != "" {
		id.Linked = b.Primary
		id.Display = b.Primary
	}
	return id
}

// Pin sets which held kasdomain is the face, or clears the display link when name is empty.
func Pin(addr, name string) error {
	addr = strings.TrimSpace(addr)
	if !strings.HasPrefix(addr, "kaspa:") {
		return fmt.Errorf("kaspa address")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Link(addr, "")
	}
	return SetPrimary(addr, name)
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
	picked, hits, err := PickOwned(addr, name)
	if err != nil {
		if len(hits) > 0 {
			return fmt.Errorf("%s (try %s)", err.Error(), strings.Join(hits, ", "))
		}
		return err
	}
	name = picked
	loadLinks()
	linkMu.Lock()
	defer linkMu.Unlock()
	if links == nil {
		links = map[string]string{}
	}
	key := strings.ToLower(addr)
	for a, n := range links {
		if n == name && a != key {
			delete(links, a)
		}
	}
	links[key] = name
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
