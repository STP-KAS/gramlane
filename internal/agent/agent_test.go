package agent

import "testing"

func TestLocalAsk(t *testing.T) {
	r, err := Ask("what is the vault bump / 234?", "")
	if err != nil {
		t.Fatal(err)
	}
	if r.Grok || r.Model != ModelLocal {
		t.Fatalf("%+v", r)
	}
	if r.Text == "" {
		t.Fatal("empty")
	}
}
