package names

import (
	"encoding/json"
	"strings"

	"gramlane/internal/quote"
)

const ResolveGrams uint64 = 12_000

// Result is a kasdomain search hit. Evidence is always indexer when Lookup worked.
type Result struct {
	Query     string          `json:"query"`
	Name      string          `json:"name"`
	Owner     string          `json:"owner,omitempty"`
	PayURI    string          `json:"payUri,omitempty"`
	AssetID   string          `json:"assetId,omitempty"`
	Evidence  string          `json:"evidence"`
	Warning   string          `json:"warning"`
	Quote     quote.Quote     `json:"quote"`
	Raw       json.RawMessage `json:"raw,omitempty"`
	Available bool            `json:"available,omitempty"`
	Linked    string          `json:"linked,omitempty"`
	Names     []string        `json:"names,omitempty"`
	Total     int             `json:"total,omitempty"`
}

func isFreeOnIndex(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "no owner") || strings.Contains(s, "404") || strings.Contains(s, "not found")
}

func looksAddr(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.HasPrefix(s, "kaspa:") || strings.HasPrefix(s, "kaspatest:")
}

// Resolve is the kasdomain hero: name or kaspa address → indexer row + gram quote.
func Resolve(raw string) (*Result, error) {
	raw = strings.TrimSpace(raw)
	qq, err := quote.Grams(ResolveGrams, "SIGN1")
	if err != nil {
		return nil, err
	}
	out := &Result{
		Query:    raw,
		Evidence: "indexer",
		Quote:    qq,
		Warning:  "Indexer first-come, first-served. Not consensus-unique. This process does not register the name.",
	}
	if looksAddr(raw) {
		id, err := ForAddress(raw)
		if err != nil {
			return out, err
		}
		out.Name = id.Display
		out.Owner = id.Address
		out.PayURI = id.Address
		out.Linked = id.Linked
		out.Names = id.Names
		out.Total = id.Total
		if id.Linked != "" {
			out.Name = id.Linked
		}
		b, _ := json.Marshal(id)
		out.Raw = b
		return out, nil
	}
	p, err := Lookup(raw)
	if p != nil {
		out.Name = p.Name
		out.Owner = p.Owner
		out.PayURI = p.PayURI
		out.AssetID = p.AssetID
		out.Raw = p.Raw
		if p.Warning != "" {
			out.Warning = p.Warning
		}
	}
	if err != nil {
		if isFreeOnIndex(err) {
			out.Available = true
			out.Warning = "No owner on the indexer. Get it once at the official shop. We host the site. No yearly rent."
			return out, nil
		}
		return out, err
	}
	return out, nil
}
