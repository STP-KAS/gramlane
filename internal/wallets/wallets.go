// Package wallets is the Kaspa wallet catalog used by KNS, Gramlane, and Kaspa Till.
// Inject connect is only claimed where a documented in-page provider exists.
package wallets

type Kind string

const (
	Hardware Kind = "hardware"
	Native   Kind = "native"
	Multi    Kind = "multi"
)

type Connect string

const (
	Inject  Connect = "inject"  // in-page provider (this origin)
	Open    Connect = "open"    // official web companion
	Install Connect = "install" // app / store / hardware; no dApp inject here
)

type Wallet struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Kind    Kind     `json:"kind"`
	Connect Connect  `json:"connect"`
	Inject  string   `json:"inject,omitempty"` // window.* key
	URL     string   `json:"url"`
	Store   string   `json:"store,omitempty"`
	Platforms []string `json:"platforms"`
	KNS     bool     `json:"kns"`
	KRC20   bool     `json:"krc20"`
	NFT     bool     `json:"nft"`
	Note    string   `json:"note"`
	Source  string   `json:"source"`
}

func All() []Wallet {
	return []Wallet{
		// Hardware
		{ID: "tangem", Name: "Tangem", Kind: Hardware, Connect: Install, URL: "https://tangem.com", Platforms: []string{"card", "mobile"}, Note: "NFC card, EAL6+ secure element, seedless multi-card backup. Kaspa since Apr 2023. Keys stay on the card — this site cannot inject it.", Source: "wiki.kaspa.org/wallet"},
		{ID: "ledger", Name: "Ledger", Kind: Hardware, Connect: Open, URL: "https://kasvault.io", Store: "https://www.ledger.com/coin/wallet/kaspa", Platforms: []string{"Nano S Plus", "Nano X", "Flex", "Stax"}, Note: "Install the Kaspa app in Ledger Live, then sign in KasVault (Chrome/Edge USB). Kastle can also link a Ledger. Ledger Live itself does not host KAS accounts.", Source: "kasvault.io; support.ledger.com"},
		{ID: "onekey", Name: "OneKey", Kind: Hardware, Connect: Install, URL: "https://onekey.so", Platforms: []string{"Classic", "Touch", "Pro", "Mini", "app", "extension"}, Note: "Open-source hardware + app with native KAS. No documented Kaspa L1 inject on this origin — use the OneKey app/device.", Source: "wiki.kaspa.org/wallet"},
		{ID: "ellipal", Name: "ELLIPAL", Kind: Hardware, Connect: Install, URL: "https://www.ellipal.com", Platforms: []string{"air-gapped", "mobile companion"}, Note: "QR air-gap. Companion app signs; this dApp never sees the key.", Source: "kaspahub.org/wallets"},
		{ID: "safepal-x1", Name: "SafePal X1", Kind: Hardware, Connect: Install, URL: "https://www.safepal.com", Store: "https://safepalsupport.zendesk.com/hc/en-us/articles/28457016304539-How-to-add-receive-or-send-KAS-Kaspa-with-SafePal-X1-Hardware-Wallet", Platforms: []string{"Bluetooth hardware"}, Note: "Native KAS on X1. Pair in the SafePal app, not here.", Source: "wiki.kaspa.org/wallet"},

		// Native / ecosystem
		{ID: "kaspium", Name: "Kaspium", Kind: Native, Connect: Install, URL: "https://kaspium.io", Store: "https://github.com/azbuky/kaspium_wallet", Platforms: []string{"iOS", "Android"}, KNS: false, Note: "Community-standard open-source mobile wallet. 12/24-word including legacy KDX/web seeds. No browser inject.", Source: "kaspium.io; wiki.kaspa.org/wallet"},
		{ID: "kasware", Name: "Kasware", Kind: Native, Connect: Inject, Inject: "kasware", URL: "https://www.kasware.xyz", Store: "https://chromewebstore.google.com/detail/hklhheigdmpoolooomdihmhlpjjdbklf", Platforms: []string{"Chrome extension", "Android APK"}, KNS: true, KRC20: true, NFT: true, Note: "window.kasware.requestAccounts(). KRC-20, KNS, NFTs, legacy 12-word. Documented dApp API.", Source: "docs.kasware.xyz"},
		{ID: "kastle", Name: "Kastle", Kind: Native, Connect: Inject, Inject: "kastle", URL: "https://kastle.cc", Store: "https://chromewebstore.google.com/detail/kastle/oambclflhjfppdmkghokjmpppmaebego", Platforms: []string{"Chrome extension", "iOS", "Android"}, KNS: true, KRC20: true, NFT: true, Note: "window.kastle.connect() then getAccount(). KRC-20, NFTs, KNS, Ledger link. L2 features exist in the wallet; this stack stays L1.", Source: "docs.kastle.cc"},
		{ID: "kaspa-ng", Name: "Kaspa NG", Kind: Native, Connect: Open, URL: "https://kaspa-ng.org", Platforms: []string{"web", "desktop"}, Note: "Successor of KDX and the legacy web wallet (wiki). Open kaspa-ng.org — it is its own wallet UI, not an inject on this origin.", Source: "wiki.kaspa.org/wallet"},
		{ID: "web-wallet", Name: "Kaspa Web Wallet", Kind: Native, Connect: Open, URL: "https://wallet.kaspanet.io", Platforms: []string{"web"}, Note: "Legacy 12-word browser wallet. Keys in local storage. Wiki: consider Kaspa NG. Still listed because it holds live funds.", Source: "wallet.kaspanet.io; wiki.kaspa.org/wallet"},
		{ID: "kdx", Name: "KDX", Kind: Native, Connect: Install, URL: "https://kdx.app", Store: "https://github.com/aspectron/kdx", Platforms: []string{"desktop node companion"}, Note: "Desktop GUI + node. Wiki: consider switching to Kaspa NG. 12-word legacy path.", Source: "kdx.app; wiki.kaspa.org/wallet"},
		{ID: "kaskeeper", Name: "KasKeeper", Kind: Native, Connect: Install, URL: "https://chromewebstore.google.com/detail/kaskeeper/bicbpicnddlclhekbmgafcbkemdikdem", Platforms: []string{"Android", "iOS", "Chrome extension"}, KRC20: true, KNS: true, Note: "KEF-backed Kaspa wallet (Tech Trends). KAS + KRC-20. No public inject API used here — install the app/extension.", Source: "Chrome Web Store; App Store KasKeeper"},
		{ID: "kurncy", Name: "Kurncy", Kind: Native, Connect: Install, URL: "https://www.kurncy.com", Platforms: []string{"iOS", "Android"}, KNS: true, KRC20: true, NFT: true, Note: "Mobile self-custody. KAS, KRC-20, KRC-721, KNS. No browser inject on this origin.", Source: "kurncy.com"},

		// Multi-chain software
		{ID: "zelcore", Name: "Zelcore", Kind: Multi, Connect: Install, URL: "https://zelcore.io", Platforms: []string{"desktop", "mobile"}, Note: "Multi-asset. Native KAS since 31 Mar 2023. Use the Zelcore app.", Source: "wiki.kaspa.org/wallet"},
		{ID: "okx", Name: "OKX Web3 Wallet", Kind: Multi, Connect: Install, URL: "https://www.okx.com/web3", Platforms: []string{"extension", "mobile"}, Note: "Non-custodial multi-chain. Holds KAS in the OKX wallet — its inject is EVM, not a Kaspa L1 provider on this page.", Source: "wiki.kaspa.org/wallet"},
		{ID: "now", Name: "NOW Wallet", Kind: Multi, Connect: Install, URL: "https://walletnow.app", Platforms: []string{"desktop", "mobile"}, Note: "Non-custodial multi-coin with built-in swap. KAS storage in the NOW app.", Source: "wiki.kaspa.org/wallet (NowWallet)"},
		{ID: "guarda", Name: "Guarda", Kind: Multi, Connect: Install, URL: "https://guarda.com", Platforms: []string{"PC", "Mac", "iOS", "Android", "web"}, Note: "Multi-currency storage/exchange. Kaspa since 12 Nov 2024.", Source: "wiki.kaspa.org/wallet"},
		{ID: "bitget", Name: "Bitget Wallet", Kind: Multi, Connect: Install, URL: "https://web3.bitget.com", Platforms: []string{"mobile", "extension"}, Note: "Multi-chain hot wallet with KAS storage. No Kaspa L1 inject here.", Source: "user list; Bitget Wallet"},
		{ID: "mathwallet", Name: "MathWallet", Kind: Multi, Connect: Install, URL: "https://mathwallet.org", Platforms: []string{"mobile", "extension"}, Note: "Multi-chain Web3. Kaspa since 24 Dec 2024. Extension is not treated as a Kaspa L1 provider on this origin.", Source: "wiki.kaspa.org/wallet"},
		{ID: "safepal-app", Name: "SafePal Software Wallet", Kind: Multi, Connect: Install, URL: "https://www.safepal.com", Platforms: []string{"mobile"}, Note: "Mobile companion to SafePal hardware. Hold/send KAS in the app.", Source: "wiki.kaspa.org/wallet"},
	}
}

func Injected() []Wallet {
	var out []Wallet
	for _, w := range All() {
		if w.Connect == Inject {
			out = append(out, w)
		}
	}
	return out
}

func ByKind(k Kind) []Wallet {
	var out []Wallet
	for _, w := range All() {
		if w.Kind == k {
			out = append(out, w)
		}
	}
	return out
}
