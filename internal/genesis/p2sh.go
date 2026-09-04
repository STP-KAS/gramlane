package genesis

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/blake2b"

	"gramlane/internal/kasaddr"
)

func artifactPath() string {
	cands := []string{Artifact}
	if wd, err := os.Getwd(); err == nil {
		cands = append(cands, filepath.Join(wd, Artifact), filepath.Join(wd, "..", "..", Artifact))
	}
	if exe, err := os.Executable(); err == nil {
		cands = append(cands, filepath.Join(filepath.Dir(exe), Artifact))
	}
	for _, p := range cands {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return Artifact
}

// compiledBlob is the portable SilAbiArtifact compiled block (silverscript#232).
type compiledBlob struct {
	Bytecode   []byte `json:"bytecode"`
	ScriptHex  string `json:"script_hex"`
	ScriptHex2 string `json:"scriptHex"`
}

func (c compiledBlob) bytes() []byte {
	if len(c.Bytecode) > 0 {
		return c.Bytecode
	}
	h := c.ScriptHex
	if h == "" {
		h = c.ScriptHex2
	}
	h = strings.TrimPrefix(strings.TrimSpace(h), "0x")
	if h == "" {
		return nil
	}
	b, err := hex.DecodeString(h)
	if err != nil {
		return nil
	}
	return b
}

func RedeemScript() ([]byte, error) {
	b, err := os.ReadFile(artifactPath())
	if err != nil {
		return nil, err
	}
	return redeemFromArtifact(b)
}

// redeemFromArtifact accepts both shapes silverc has emitted:
// map keyed by contract name (v1-rc1 WorkCredit-live.json) and the
// array-of-contracts example in silverscript#232.
func redeemFromArtifact(b []byte) ([]byte, error) {
	var mapped struct {
		Contracts map[string]struct {
			Compiled compiledBlob `json:"compiled"`
		} `json:"contracts"`
	}
	if json.Unmarshal(b, &mapped) == nil {
		if c, ok := mapped.Contracts["WorkCredit"]; ok {
			if bc := c.Compiled.bytes(); len(bc) > 0 {
				return bc, nil
			}
		}
	}
	var listed struct {
		Contracts []struct {
			Name     string       `json:"name"`
			Compiled compiledBlob `json:"compiled"`
		} `json:"contracts"`
	}
	if json.Unmarshal(b, &listed) == nil {
		for _, c := range listed.Contracts {
			if c.Name == "WorkCredit" {
				if bc := c.Compiled.bytes(); len(bc) > 0 {
					return bc, nil
				}
			}
		}
	}
	return nil, os.ErrNotExist
}

func P2SH() (addr, hashHex string, redeemLen int, err error) {
	redeem, err := RedeemScript()
	if err != nil {
		return "", "", 0, err
	}
	sum := blake2b.Sum256(redeem)
	addr = kasaddr.Encode("kaspa", sum[:], kasaddr.VersionScriptHash)
	return addr, hex.EncodeToString(sum[:]), len(redeem), nil
}
