// Package genesis is the WorkCredit instance tied to the 0.5 KAS L1 sale.
// Template hash is the code. Constructor state is this issuer/holder/credits/lane.
// Deploy is sendKaspa to the P2SH of the compiled redeem script.
package genesis

import (
	"encoding/hex"
	"os"
	"strconv"

	"gramlane/internal/chain"
)

const (
	SaleTx          = "c1799b0de40f71cfd7a153684ef22326ad920d0dca2a8b519ce2c8379c4f7bc2"
	Explorer        = "https://explorer.kaspa.org/txs/" + SaleTx
	DeskAddress     = "kaspa:qpjm8kzpcj5he3hg9msrdc78a3k46zda866pucwetprgtgc7s3ry2kq38atpq"
	HolderAddress   = "kaspa:qrw0tuxnjukkrk7lm5v7chx4qpu8f8q0janp7jdk80z4tc8anu62g62s4zsj0"
	IssuerPubHex    = "65b3d841c4a97cc6e82ee036e3c7ec6d5d09bd3eb41e61d9584685a31e844645"
	HolderPubHex    = "dcf5f0d3972d61dbdfdd19ec5cd50078749c0f97661f49b63bc555e0fd9f34a4"
	Credits         = 500_000
	Lane            = "SEQ1"
	SaleSompi       = 50_000_000
	TemplateHashHex = "c61458da2134faad500c05cc35fb5dfd2db7e44538782703207f14926a11521e"
	Artifact        = "contracts/v1/WorkCredit-live.json"
)

type Plan struct {
	SaleTx         string `json:"saleTx"`
	Explorer       string `json:"explorer"`
	Desk           string `json:"desk"`
	Holder         string `json:"holder"`
	IssuerPub      string `json:"issuerPubXOnly"`
	HolderPub      string `json:"holderPubXOnly"`
	Credits        int    `json:"credits"`
	Lane           string `json:"lane"`
	SaleKAS        string `json:"saleKas"`
	TemplateHash   string `json:"templateHash"`
	Artifact       string `json:"artifact"`
	OnChainSale    bool   `json:"onChainSale"`
	VoucherOnChain bool   `json:"voucherOnChain"`
	Note           string `json:"note"`
	CtorPath       string `json:"ctorPath"`
	P2SH           string `json:"p2sh"`
	ScriptHash     string `json:"scriptHash"`
	RedeemLen      int    `json:"redeemLen"`
	FundSompi      uint64 `json:"fundSompi"`
	FundKAS        string `json:"fundKas"`
	VoucherTx      string `json:"voucherTx,omitempty"`
	VoucherIdx     int    `json:"voucherIndex"`
}

func Live() Plan {
	art := Artifact
	if _, err := os.Stat(art); err != nil {
		art = ""
	}
	p := Plan{
		SaleTx:         SaleTx,
		Explorer:       Explorer,
		Desk:           DeskAddress,
		Holder:         HolderAddress,
		IssuerPub:      IssuerPubHex,
		HolderPub:      HolderPubHex,
		Credits:        Credits,
		Lane:           Lane,
		SaleKAS:        "0.5",
		TemplateHash:   TemplateHashHex,
		Artifact:       art,
		OnChainSale:    true,
		VoucherOnChain: false,
		CtorPath:       "contracts/v1/ctor-WorkCredit-live.json",
		FundSompi:      SaleSompi,
		FundKAS:        "0.5",
		Note:           "Sale is on L1. Deploy = Kasware sendKaspa of 0.5 KAS to the P2SH of this redeem script. That output is the WorkCredit UTXO. consume() is a later spend of that UTXO.",
	}
	if addr, hash, n, err := P2SH(); err == nil {
		p.P2SH = addr
		p.ScriptHash = hash
		p.RedeemLen = n
		if utxos, err := chain.AddressUTXOs(addr); err == nil && len(utxos) > 0 {
			p.VoucherOnChain = true
			p.VoucherTx = utxos[0].TxID
			p.VoucherIdx = utxos[0].Index
			p.Note = "P2SH UTXO is on L1 (" + utxos[0].TxID + ":" + strconv.Itoa(utxos[0].Index) + "). 500000 grams prepaid. Jobs burn grams on the sequencer ledger. consume() of this UTXO is still a later spend."
		}
	} else {
		p.Note += " P2SH encode failed: " + err.Error()
	}
	return p
}

func MustHex(h string) []byte {
	b, err := hex.DecodeString(h)
	if err != nil {
		panic(err)
	}
	return b
}
