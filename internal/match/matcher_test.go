package match

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/retronian/retronian-scraper/internal/db"
	"github.com/retronian/retronian-scraper/internal/scan"
)

func TestMatcher_NameFallback(t *testing.T) {
	games := []db.Game{
		{
			ID: "hillsfar",
			ROMs: []db.ROM{
				{
					Name: "Advanced Dungeons & Dragons - Hillsfar (Japan)",
					SHA1: "5b212ee25c75449e39fa97f982cc44a3068c358d",
				},
			},
		},
	}

	got := New(games).Match(
		"/roms/Advanced Dungeons & Dragons - Hillsfar (Japan).nes",
		scan.Hashes{SHA1: "5bc6b4d5e2b27983e678376a95b041ecb3abe819"},
	)
	if got.Tier != TierNameFallback {
		t.Fatalf("tier: want name fallback, got %v", got.Tier)
	}
	if got.Game == nil || got.Game.ID != "hillsfar" {
		t.Fatalf("game: got %#v", got.Game)
	}
}

func TestMatcher_NameFallbackUsesZipInnerName(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "outer-name.zip")
	if err := writeZip(path, "Advanced Dungeons & Dragons - Hillsfar (Japan).nes", []byte("rom")); err != nil {
		t.Fatal(err)
	}
	games := []db.Game{
		{
			ID: "hillsfar",
			ROMs: []db.ROM{
				{Name: "Advanced Dungeons & Dragons - Hillsfar (Japan)"},
			},
		},
	}

	got := New(games).Match(path, scan.Hashes{})
	if got.Tier != TierNameFallback {
		t.Fatalf("tier: want name fallback, got %v", got.Tier)
	}
	if got.Game == nil || got.Game.ID != "hillsfar" {
		t.Fatalf("game: got %#v", got.Game)
	}
}

func TestMatcher_ZipOuterNameIgnored(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Advanced Dungeons & Dragons - Hillsfar (Japan).zip")
	if err := writeZip(path, "different-game.nes", []byte("rom")); err != nil {
		t.Fatal(err)
	}
	games := []db.Game{
		{
			ID: "hillsfar",
			ROMs: []db.ROM{
				{Name: "Advanced Dungeons & Dragons - Hillsfar (Japan)"},
			},
		},
	}

	got := New(games).Match(path, scan.Hashes{})
	if got.Tier != TierNone {
		t.Fatalf("tier: want none because zip outer name must be ignored, got %v", got.Tier)
	}
}

func TestMatcher_HashBeatsNameFallback(t *testing.T) {
	games := []db.Game{
		{
			ID: "hash-match",
			ROMs: []db.ROM{
				{Name: "Wrong Name", SHA1: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			},
		},
		{
			ID: "name-match",
			ROMs: []db.ROM{
				{Name: "Example Game (Japan)", SHA1: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
			},
		},
	}

	got := New(games).Match(
		"/roms/Example Game (Japan).zip",
		scan.Hashes{SHA1: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	)
	if got.Tier != TierSHA1 {
		t.Fatalf("tier: want sha1, got %v", got.Tier)
	}
	if got.Game == nil || got.Game.ID != "hash-match" {
		t.Fatalf("game: got %#v", got.Game)
	}
}

func TestMatcher_NameFallbackIgnoresTrailingRegionTag(t *testing.T) {
	games := []db.Game{
		{
			ID: "brightis",
			ROMs: []db.ROM{
				{Name: "Brightis (Japan)"},
			},
		},
	}

	got := New(games).Match("/roms/Brightis.chd", scan.Hashes{})
	if got.Tier != TierNameFallback {
		t.Fatalf("tier: want name fallback, got %v", got.Tier)
	}
	if got.Game == nil || got.Game.ID != "brightis" {
		t.Fatalf("game: got %#v", got.Game)
	}
}

func TestMatcher_NameFallbackUsesGameTitles(t *testing.T) {
	games := []db.Game{
		{
			ID: "addie",
			Titles: []db.Title{
				{Text: "Addie No Okurimono: To Moze From Addie"},
			},
		},
	}

	got := New(games).Match("/roms/Addie no Okurimono - To Moze from Addie.chd", scan.Hashes{})
	if got.Tier != TierNameFallback {
		t.Fatalf("tier: want name fallback, got %v", got.Tier)
	}
	if got.Game == nil || got.Game.ID != "addie" {
		t.Fatalf("game: got %#v", got.Game)
	}
}

func writeZip(path, name string, data []byte) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	zw := zip.NewWriter(out)
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	return zw.Close()
}
