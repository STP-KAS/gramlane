// Package wallets is a catalog of official Kaspa wallets.
// This desk does not ship in-page inject.
package wallets

type Kind string

const (
	Hardware Kind = "hardware"
	Native   Kind = "native"
	Multi    Kind = "multi"
)

type Connect string

const (
	Inject  Connect = "inject"
	Open    Connect = "open"
	Install Connect = "install"
)

type Wallet struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Kind      Kind     `json:"kind"`
	Connect   Connect  `json:"connect"`
	Inject    string   `json:"inject,omitempty"`
	URL       string   `json:"url"`
	Store     string   `json:"store,omitempty"`
	Platforms []string `json:"platforms"`
	KNS       bool     `json:"kns"`
	KRC20     bool     `json:"krc20"`
	NFT       bool     `json:"nft"`
	Note      string   `json:"note"`
	Source    string   `json:"source"`
}

func All() []Wallet {
	return []Wallet{
		{ID: "kaspium", Name: "Kaspium", Kind: Native, Connect: Install, URL: "https://kaspium.io", Platforms: []string{"iOS", "Android"}, Note: "Mobile. No browser inject."},
		{ID: "kasware", Name: "Kasware", Kind: Native, Connect: Open, URL: "https://www.kasware.xyz", Platforms: []string{"Chrome extension"}, Note: "Official site. This desk does not ship inject."},
		{ID: "kastle", Name: "Kastle", Kind: Native, Connect: Open, URL: "https://kastle.cc", Platforms: []string{"Chrome extension"}, Note: "Official site. This desk does not ship inject."},
		{ID: "kaspa-ng", Name: "Kaspa NG", Kind: Native, Connect: Open, URL: "https://kaspa-ng.org", Platforms: []string{"web", "desktop"}, Note: "Own wallet UI."},
	}
}

func Injected() []Wallet { return nil }

func ByKind(k Kind) []Wallet {
	var out []Wallet
	for _, w := range All() {
		if w.Kind == k {
			out = append(out, w)
		}
	}
	return out
}
