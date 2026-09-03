package genesis

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

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

func RedeemScript() ([]byte, error) {
	b, err := os.ReadFile(artifactPath())
	if err != nil {
		return nil, err
	}
	var art struct {
		Contracts map[string]struct {
			Compiled struct {
				Bytecode []byte `json:"bytecode"`
			} `json:"compiled"`
		} `json:"contracts"`
	}
	if err := json.Unmarshal(b, &art); err != nil {
		return nil, err
	}
	c, ok := art.Contracts["WorkCredit"]
	if !ok || len(c.Compiled.Bytecode) == 0 {
		return nil, os.ErrNotExist
	}
	return c.Compiled.Bytecode, nil
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
