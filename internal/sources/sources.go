// Package sources is the pin list this desk actually uses.
// Short local URLs redirect to the canonical GitHub / tracker.
package sources

type Pin struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Local string `json:"local,omitempty"`
	Chip  string `json:"chip"`
	Use   string `json:"use"`
}

// Used is every source this dApp cites. Refuse-rows stay on /sources so they are not a pin.
var Used = []Pin{
	{ID: "kips", Title: "kaspanet/kips", URL: "https://github.com/kaspanet/kips", Local: "/kips-repo", Chip: "Active", Use: "KIP repository. Law is merged Active only."},
	{ID: "kip-16", Title: "KIP-16 ZK precompile", URL: "https://github.com/kaspanet/kips/blob/master/kip-0016.md", Local: "/kip-16", Chip: "Active", Use: "Toccata. Covenants++ opcode."},
	{ID: "kip-17", Title: "KIP-17 covenants", URL: "https://github.com/kaspanet/kips/blob/master/kip-0017.md", Local: "/kip-17", Chip: "Active", Use: "KasName and WorkCredit spend rules."},
	{ID: "kip-20", Title: "KIP-20 covenant IDs", URL: "https://github.com/kaspanet/kips/blob/master/kip-0020.md", Local: "/kip-20", Chip: "Active", Use: "Outpoint id. The string bakery is not unique."},
	{ID: "kip-21", Title: "KIP-21 sequencing / grams", URL: "https://github.com/kaspanet/kips/blob/master/kip-0021.md", Local: "/kip-21", Chip: "Active", Use: "Mass in grams. 50 lanes/block, 1e9 gas/lane on L1. Desk MSG1 is not a subnetwork_id."},
	{ID: "kccs", Title: "kaspanet/kccs", URL: "https://github.com/kaspanet/kccs", Local: "/kccs", Chip: "Draft", Use: "Read so we can refuse a token. WorkCredit is the voucher."},
	{ID: "kccs-20", Title: "kccs#20 vectors", URL: "https://github.com/kaspanet/kccs/pull/20", Local: "/kccs-20", Chip: "open", Use: "KCC-0020 still Draft. Not GRAM."},
	{ID: "silverscript", Title: "silverscript releases", URL: "https://github.com/kaspanet/silverscript/releases", Local: "/silverscript", Chip: "v1-rc1", Use: "Official compiler list. Only v1-rc1 as of 5 Sep 2026."},
	{ID: "v1-rc1", Title: "silverc v1-rc1", URL: "https://github.com/kaspanet/silverscript/releases/tag/v1-rc1", Local: "/v1-rc1", Chip: "pin", Use: "KasName.sil and WorkCredit.sil. Not master."},
	{ID: "ss-232", Title: "silverscript#232 ABI", URL: "https://github.com/kaspanet/silverscript/pull/232", Local: "/ss-232", Chip: "merged", Use: "Portable ABI JSON this desk reads."},
	{ID: "ss-234", Title: "silverscript#234 framing", URL: "https://github.com/kaspanet/silverscript/pull/234", Local: "/ss-234", Chip: "closed", Use: "Foreign readInputState. Closed unmerged. Live /234."},
	{ID: "ss-243", Title: "silverscript#243 budget", URL: "https://github.com/kaspanet/silverscript/issues/243", Local: "/ss-243", Chip: "open", Use: "consume() has no compute budget in the artifact."},
	{ID: "explained-kips", Title: "Kaspa Explained KIPs", URL: "https://kaspaexplained.com/kips", Local: "/explained", Chip: "tracker", Use: "Human map. Not protocol law."},
	{ID: "toccata-status", Title: "Toccata status", URL: "https://kaspaexplained.com/toccata-status", Local: "/toccata-status", Chip: "tracker", Use: "What is live on L1."},
	{ID: "explorer", Title: "Kaspa explorer", URL: "https://explorer.kaspa.org", Local: "/explorer", Chip: "live", Use: "L1 tx links."},
	{ID: "api-kaspa", Title: "api.kaspa.org", URL: "https://api.kaspa.org", Local: "/api-kaspa", Chip: "live", Use: "DAA / balance probes. Not kasdomain."},
	{ID: "wiki-wallet", Title: "wiki.kaspa.org/wallet", URL: "https://wiki.kaspa.org/wallet", Local: "/wiki-wallet", Chip: "wiki", Use: "Wallet list. Not this desk."},
	{ID: "tedx-prior", Title: "TEDx timestamp idea", URL: "https://www.youtube.com/watch?v=pA19Tf5wFEA", Local: "/tedx", Chip: "idea", Use: "Prior art: hash on a public ledger. Not USPTO."},
	{ID: "stppstp", Title: "@StppStp", URL: "https://x.com/StppStp", Local: "/x", Chip: "human", Use: "Only X for this project. This site never DMs you."},
	{ID: "gramlane", Title: "STP-KAS/gramlane", URL: "https://github.com/STP-KAS/gramlane", Local: "/repo", Chip: "this", Use: "This dApp."},
	{ID: "delusional", Title: "project delusional", URL: "https://github.com/STP-KAS/project-delusional", Local: "/delusional", Chip: "index", Use: "Stack index."},
	{ID: "masterfile", Title: "kaspa-master-file", URL: "https://github.com/STP-KAS/kaspa-master-file", Local: "/masterfile-repo", Chip: "this", Use: "Public Kaspa map. Tab /masterfile."},
}

var Refuse = []Pin{
	{ID: "kip-2", Title: "KIP-2 DAGKnight", URL: "https://github.com/kaspanet/kips/blob/master/kip-0002.md", Chip: "wrong", Use: "Proposed. Not in this dApp."},
	{ID: "coderofstuff", Title: "coderofstuff forks / dk-wiki", URL: "https://github.com/coderofstuff/dk-wiki", Chip: "wrong", Use: "Unofficial. Not a pin."},
}

func ByLocal(path string) (Pin, bool) {
	for _, p := range Used {
		if p.Local == path {
			return p, true
		}
	}
	return Pin{}, false
}
