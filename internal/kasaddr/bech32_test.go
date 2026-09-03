package kasaddr

import (
	"encoding/hex"
	"testing"
)

func TestEncodeKnownPubKey(t *testing.T) {
	payload, err := hex.DecodeString("65b3d841c4a97cc6e82ee036e3c7ec6d5d09bd3eb41e61d9584685a31e844645")
	if err != nil {
		t.Fatal(err)
	}
	got := Encode("kaspa", payload, VersionPubKey)
	want := "kaspa:qpjm8kzpcj5he3hg9msrdc78a3k46zda866pucwetprgtgc7s3ry2kq38atpq"
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestRoundTripPubKey(t *testing.T) {
	payload, _ := hex.DecodeString("65b3d841c4a97cc6e82ee036e3c7ec6d5d09bd3eb41e61d9584685a31e844645")
	addr := Encode("kaspa", payload, VersionPubKey)
	got, ver, err := DecodePayload(addr)
	if err != nil {
		t.Fatal(err)
	}
	if ver != VersionPubKey {
		t.Fatal(ver)
	}
	if hex.EncodeToString(got[:32]) != hex.EncodeToString(payload) {
		t.Fatalf("%x vs %x", got, payload)
	}
}
