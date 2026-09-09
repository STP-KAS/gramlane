package masterfile

import "testing"

func TestLiveHasKIP21AndDiscord(t *testing.T) {
	f := Live()
	if f.Title == "" || f.Updated == "" || len(f.Sections) < 4 {
		t.Fatalf("%+v", f)
	}
	var kip, disc, tg, vision, explained, video, research, kurrent, sil, mix, darwin bool
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
			if r.URL == "https://research.kas.pa" {
				research = true
			}
			if r.URL == "https://research.kas.pa/t/kurrent-an-eltoo-inspired-latest-state-channel-on-kaspa/494" {
				kurrent = true
			}
			if r.URL == "https://github.com/kaspanet/silverscript/releases/tag/v1.0.0" {
				sil = true
			}
			if r.URL == "https://github.com/STP-KAS/mixer-concept" {
				mix = true
			}
			if r.URL == "https://github.com/STP-KAS/gramlanepeglab" {
				darwin = true
			}
		}
	}
	if f.Updated < "2026-09-09" {
		t.Fatalf("stale master file %s", f.Updated)
	}
	if !kip || !disc || !tg || !vision || !explained || !video || !research || !kurrent || !sil || !mix || !darwin {
		t.Fatalf("kip=%v disc=%v tg=%v vision=%v explained=%v video=%v research=%v kurrent=%v sil=%v mix=%v darwin=%v", kip, disc, tg, vision, explained, video, research, kurrent, sil, mix, darwin)
	}
}
