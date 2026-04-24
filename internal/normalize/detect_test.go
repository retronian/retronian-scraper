package normalize

import (
	"os"
	"path/filepath"
	"testing"
)

func writeROMs(t *testing.T, dir string, files []string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		p := filepath.Join(dir, f)
		if err := os.WriteFile(p, []byte("dummy"), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}
}

func TestDetectPlatform_AliasHit(t *testing.T) {
	dir := t.TempDir()
	d, err := DetectPlatform(dir, "GameBoy")
	if err != nil {
		t.Fatal(err)
	}
	if d.InternalID != "gb" || d.Method != DetectByAlias {
		t.Errorf("want (gb, alias), got (%q, %q)", d.InternalID, d.Method)
	}
}

func TestDetectPlatform_ContentsHit_UnambiguousExtension(t *testing.T) {
	dir := t.TempDir()
	writeROMs(t, dir, []string{"a.nes", "b.nes", "c.fds"})

	d, err := DetectPlatform(dir, "Misc-Unknown-Folder")
	if err != nil {
		t.Fatal(err)
	}
	if d.InternalID != "fc" || d.Method != DetectByContents {
		t.Errorf("want (fc, contents), got (%q, %q)", d.InternalID, d.Method)
	}
	if d.Score != 3 {
		t.Errorf("score: want 3, got %d", d.Score)
	}
}

func TestDetectPlatform_ContentsTieBreak_BinFavorsPS1(t *testing.T) {
	// .bin belongs to both md and ps1. Tie-break order makes ps1 win.
	dir := t.TempDir()
	writeROMs(t, dir, []string{"a.bin", "b.bin"})

	d, err := DetectPlatform(dir, "ambiguous")
	if err != nil {
		t.Fatal(err)
	}
	if d.InternalID != "ps1" {
		t.Errorf("want ps1 (tie-break winner), got %q", d.InternalID)
	}
}

func TestDetectPlatform_EmptyFolder(t *testing.T) {
	dir := t.TempDir()
	d, err := DetectPlatform(dir, "WeirdName")
	if err != nil {
		t.Fatal(err)
	}
	if d.Method != DetectFailed {
		t.Errorf("want failed, got %q (id=%q)", d.Method, d.InternalID)
	}
}

func TestDetectPlatform_AliasBeatsContents(t *testing.T) {
	// Folder is named "Game Boy" but contains .sfc files; alias should win.
	dir := t.TempDir()
	writeROMs(t, dir, []string{"a.sfc", "b.sfc"})

	d, err := DetectPlatform(dir, "Game Boy")
	if err != nil {
		t.Fatal(err)
	}
	if d.InternalID != "gb" || d.Method != DetectByAlias {
		t.Errorf("want (gb, alias), got (%q, %q)", d.InternalID, d.Method)
	}
}

func TestDetectPlatform_ContentsRecursive(t *testing.T) {
	// Walk recursively into subfolders.
	dir := t.TempDir()
	sub := filepath.Join(dir, "subdir")
	writeROMs(t, sub, []string{"a.gba"})

	d, err := DetectPlatform(dir, "RandomName")
	if err != nil {
		t.Fatal(err)
	}
	if d.InternalID != "gba" {
		t.Errorf("want gba, got %q", d.InternalID)
	}
}

func TestDetectPlatform_IgnoresUnknownExtensions(t *testing.T) {
	dir := t.TempDir()
	writeROMs(t, dir, []string{"a.txt", "b.png", "c.jpg"})

	d, err := DetectPlatform(dir, "RandomName")
	if err != nil {
		t.Fatal(err)
	}
	if d.Method != DetectFailed {
		t.Errorf("want failed for non-ROM files, got %q", d.Method)
	}
}
