package normalize

import (
	"sort"
	"testing"

	"github.com/retronian/retronian-scraper/internal/scan"
)

func TestLookupProfile_CaseInsensitive(t *testing.T) {
	cases := []string{"es-de", "ES-DE", " Es-De ", "minui", "MINUI"}
	for _, in := range cases {
		if _, err := LookupProfile(in); err != nil {
			t.Errorf("LookupProfile(%q) unexpected error: %v", in, err)
		}
	}
}

func TestLookupProfile_Unknown(t *testing.T) {
	if _, err := LookupProfile("retroarch"); err == nil {
		t.Errorf("LookupProfile(retroarch) expected error, got nil")
	}
}

func TestProfile_TargetFolder(t *testing.T) {
	esde := Profiles[FrontendESDE]
	if got, ok := esde.TargetFolder("gb"); !ok || got != "gb" {
		t.Errorf("es-de gb: want (gb,true), got (%q,%v)", got, ok)
	}
	if got, ok := esde.TargetFolder("ps1"); !ok || got != "psx" {
		t.Errorf("es-de ps1: want (psx,true), got (%q,%v)", got, ok)
	}

	minui := Profiles[FrontendMinUI]
	if got, ok := minui.TargetFolder("n64"); ok || got != "n64" {
		t.Errorf("minui n64 fallback: want (n64,false), got (%q,%v)", got, ok)
	}
	if got, ok := minui.TargetFolder("gb"); !ok || got != "Game Boy (GB)" {
		t.Errorf("minui gb: want (Game Boy (GB),true), got (%q,%v)", got, ok)
	}
}

// All known internal platforms must resolve via TargetFolder for every
// frontend (either officially or via fallback). This guards against
// typos in the Folders maps when adding new platforms.
func TestProfiles_CoverAllInternalPlatforms(t *testing.T) {
	known := scan.KnownPlatforms()
	sort.Strings(known)

	for id, p := range Profiles {
		for _, ip := range known {
			got, _ := p.TargetFolder(ip)
			if got == "" {
				t.Errorf("frontend %s: TargetFolder(%q) returned empty name", id, ip)
			}
		}
	}
}

func TestKnownFrontends_Sorted(t *testing.T) {
	got := KnownFrontends()
	want := []string{"batocera", "es-de", "minui", "onion", "recalbox", "unuui"}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: want %d, got %d (%v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] want %q, got %q", i, want[i], got[i])
		}
	}
}
