package gramchat

import "testing"

func TestRoomAndPut(t *testing.T) {
	ResetForTest(t.TempDir())
	if Room("Bakery") != "bakery.kas" || Room("#board") != "board" {
		t.Fatal(Room("Bakery"), Room("#board"))
	}
	if _, err := Put("bakery", "kaspa:q", "aa", "bb", 1); err == nil {
		t.Fatal("short nonce")
	}
	nonce := "00112233445566778899aabb"
	box := "00112233445566778899aabbccddeeff00112233"
	n, err := Put("bakery", "kaspa:qtest", nonce, box, 9000)
	if err != nil || n.Room != "bakery.kas" || n.Grams != 9000 {
		t.Fatalf("%v %+v", err, n)
	}
	got := List("bakery.kas")
	if len(got) != 1 || got[0].Box != box {
		t.Fatalf("%+v", got)
	}
	if len(List("other.kas")) != 0 {
		t.Fatal("leak")
	}
}
