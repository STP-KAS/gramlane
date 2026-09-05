// Package masterfile is the public Kaspa pin list (not Kaspa core).
package masterfile

import (
	_ "embed"
	"encoding/json"
)

//go:embed master.json
var raw []byte

type Row struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Chip string `json:"chip"`
	Note string `json:"note"`
}

type Section struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Rows  []Row  `json:"rows"`
}

type File struct {
	Title      string    `json:"title"`
	Repo       string    `json:"repo"`
	Updated    string    `json:"updated"`
	Disclaimer string    `json:"disclaimer"`
	Sections   []Section `json:"sections"`
}

func Live() File {
	var f File
	_ = json.Unmarshal(raw, &f)
	return f
}
