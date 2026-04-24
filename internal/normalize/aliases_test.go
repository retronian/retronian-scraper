package normalize

import "testing"

func TestNormalizeFolderName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Game Boy", "gameboy"},
		{"game-boy", "gameboy"},
		{"GameBoy", "gameboy"},
		{"GAME_BOY", "gameboy"},
		{"  gameboy  ", "gameboy"},
		{"Game.Boy", "gameboy"},
		{"Game Boy (GB)", "gameboygb"},
		{"Sony PlayStation (PS)", "sonyplaystationps"},
		{"ｓｆｃ", "sfc"}, // NFKC: full-width to half-width
		{"", ""},
		{"---", ""},
	}
	for _, c := range cases {
		if got := NormalizeFolderName(c.in); got != c.want {
			t.Errorf("NormalizeFolderName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLookupByAlias_AllInternalIDs(t *testing.T) {
	for id := range aliasTable {
		got, ok := LookupByAlias(id)
		if !ok || got != id {
			t.Errorf("LookupByAlias(%q) = (%q, %v), want (%q, true)", id, got, ok, id)
		}
	}
}

func TestLookupByAlias_FrontendOfficialNames(t *testing.T) {
	// Every frontend's official folder name must round-trip back to the
	// internal ID via NormalizeFolderName + LookupByAlias.
	for fid, p := range Profiles {
		for internalID, folder := range p.Folders {
			canonical := NormalizeFolderName(folder)
			got, ok := LookupByAlias(canonical)
			if !ok {
				t.Errorf("frontend %s: official folder %q (canonical %q) for %q is not in aliasTable",
					fid, folder, canonical, internalID)
				continue
			}
			if got != internalID {
				t.Errorf("frontend %s: official folder %q (canonical %q) for %q resolved to %q",
					fid, folder, canonical, internalID, got)
			}
		}
	}
}

func TestLookupByAlias_Miss(t *testing.T) {
	cases := []string{"", "retroarch", "switch", "ps2", "n65"}
	for _, c := range cases {
		if got, ok := LookupByAlias(c); ok {
			t.Errorf("LookupByAlias(%q) = (%q, true), want miss", c, got)
		}
	}
}
