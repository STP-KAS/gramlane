package framing

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCanonicalIs42AndVaultReadsOne(t *testing.T) {
	b := CanonicalEncode(Amount1, DemoOwner())
	if len(b) != TokenStateLen {
		t.Fatalf("len %d", len(b))
	}
	n, ok := VaultAmount(b)
	if !ok || n != 1 {
		t.Fatalf("vault %d ok=%v", n, ok)
	}
	if err := PinHeaders(b); err != nil {
		t.Fatal(err)
	}
}

func TestAttackSameLengthVaultReads264(t *testing.T) {
	b := AttackEncode(Amount1, DemoOwner())
	if len(b) != len(CanonicalEncode(Amount1, DemoOwner())) {
		t.Fatal("attack must preserve length")
	}
	n, ok := VaultAmount(b)
	if !ok || n != VaultSeesForAmount1 {
		t.Fatalf("vault %d want %d", n, VaultSeesForAmount1)
	}
	want := binary.LittleEndian.Uint64([]byte{0x08, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if n != want {
		t.Fatalf("window %d want %d", n, want)
	}
	if err := PinHeaders(b); err == nil {
		t.Fatal("pin must reject the reframe")
	}
}

func TestHugeReframeIsTwoToTheFiftyNinePlusEight(t *testing.T) {
	b := AttackEncode(HugeReal, DemoOwner())
	n, ok := VaultAmount(b)
	if !ok || n != HugeVault {
		t.Fatalf("vault %d want %d", n, HugeVault)
	}
}

func TestContractSourcesDoNotReadForeignState(t *testing.T) {
	root := filepath.Join("..", "..", "contracts")
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".sil") {
			return err
		}
		if strings.Contains(info.Name(), "FORBIDDEN") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(b), "readInputState") {
			t.Errorf("%s uses readInputState — forbidden on v1-rc1 (#234)", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
