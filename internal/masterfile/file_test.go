package masterfile

import "testing"

func TestLiveHasKIP21AndDiscord(t *testing.T) {
	f := Live()
	if f.Title == "" || f.Updated == "" || len(f.Sections) < 4 {
		t.Fatalf("%+v", f)
	}
	var kip, disc, tg, vision, explained, video bool
	for _, s := range f.Sections {
		for _, r := range s.Rows {
			if r.Name == "project delusional" {
				vision = true
			}
			if r.URL == "https://kaspaexplained.com/start-here" {
				explained = true
			}
			if r.URL == "https://x.com/kaspaunchained/status/2096211914825285808" {
				video = true
			}
			if r.Name == "KIP-21" {
				kip = true
			}
			if r.Name == "Discord invite" {
				disc = true
			}
			if r.Name == "Telegram Bot API" {
				tg = true
			}
		}
	}
	if !kip || !disc || !tg || !vision || !explained || !video {
		t.Fatalf("kip=%v disc=%v tg=%v vision=%v explained=%v video=%v", kip, disc, tg, vision, explained, video)
	}
}
