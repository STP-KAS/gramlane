package httpcache

import "testing"

func TestGetCaches(t *testing.T) {
	ResetForTest()
	n := 0
	fn := func() ([]byte, error) {
		n++
		return []byte("ok"), nil
	}
	a, err := Get("k", fn)
	if err != nil || string(a) != "ok" || n != 1 {
		t.Fatalf("first %v %s n=%d", err, a, n)
	}
	b, err := Get("k", fn)
	if err != nil || string(b) != "ok" || n != 1 {
		t.Fatalf("cached %v n=%d", err, n)
	}
}
