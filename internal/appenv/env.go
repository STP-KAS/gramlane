// Package appenv is listen address, public URL, and data dir for a small VPS.
package appenv

import (
	"os"
	"path/filepath"
	"strings"
)

func Listen() string {
	if a := strings.TrimSpace(os.Getenv("LISTEN")); a != "" {
		return a
	}
	if a := strings.TrimSpace(os.Getenv("GRAMLANE_ADDR")); a != "" {
		return a
	}
	return "127.0.0.1:8081"
}

func PublicBase() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_BASE")), "/")
}

func DataDir() string {
	if d := strings.TrimSpace(os.Getenv("DATA_DIR")); d != "" {
		return d
	}
	if d := strings.TrimSpace(os.Getenv("GRAMLANE_DATA")); d != "" {
		return d
	}
	return "data"
}

func File(name string) string {
	return filepath.Join(DataDir(), name)
}

func PublicHost() bool {
	return PublicBase() != ""
}
