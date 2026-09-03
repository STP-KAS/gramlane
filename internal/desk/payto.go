package desk

import (
	"os"
	"strings"
)

const Env = "GRAMLANE_PAYTO"

// PayTo is the KAS fallback destination. It must not be the connected Kasware
// account — the wallet warns on self-send. Set GRAMLANE_PAYTO or payto.txt.
func PayTo() string {
	if v := strings.TrimSpace(os.Getenv(Env)); v != "" {
		return v
	}
	b, err := os.ReadFile("payto.txt")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
}
