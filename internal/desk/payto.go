package desk

import (
	"os"
	"strings"
)

const Env = "GRAMLANE_PAYTO"

// DefaultPayTo is the desk that received the 0.5 KAS sale
// (txid c1799b0de40f71cfd7a153684ef22326ad920d0dca2a8b519ce2c8379c4f7bc2).
const DefaultPayTo = "kaspa:qpjm8kzpcj5he3hg9msrdc78a3k46zda866pucwetprgtgc7s3ry2kq38atpq"

// PayTo is the KAS fallback destination. Must not be the connected Kasware
// account. Override with GRAMLANE_PAYTO or payto.txt.
func PayTo() string {
	if v := strings.TrimSpace(os.Getenv(Env)); v != "" {
		return v
	}
	b, err := os.ReadFile("payto.txt")
	if err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			return line
		}
	}
	return DefaultPayTo
}
