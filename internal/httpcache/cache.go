// Package httpcache keeps indexer and node JSON for a short TTL.
package httpcache

import (
	"sync"
	"time"
)

const TTL = 45 * time.Second

type entry struct {
	at   time.Time
	body []byte
	err  error
}

var (
	mu sync.Mutex
	m  = map[string]entry{}
)

func Get(key string, fetch func() ([]byte, error)) ([]byte, error) {
	now := time.Now()
	mu.Lock()
	if e, ok := m[key]; ok && now.Sub(e.at) < TTL {
		mu.Unlock()
		return e.body, e.err
	}
	mu.Unlock()
	body, err := fetch()
	mu.Lock()
	m[key] = entry{at: time.Now(), body: body, err: err}
	mu.Unlock()
	return body, err
}

func ResetForTest() {
	mu.Lock()
	m = map[string]entry{}
	mu.Unlock()
}
