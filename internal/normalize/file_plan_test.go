package normalize

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/retronian/retronian-scraper/internal/db"
	"github.com/retronian/retronian-scraper/internal/match"
)

func matchedFile(path string) match.Result {
	game := &db.Game{
		ID: "super-mario-land",
		Titles: []db.Title{
			{Text: "Super Mario Land", Lang: "en", Verified: true},
			{Text: "スーパーマリオランド", Lang: "ja", Script: "Jpan", Verified: true},
		},
	}
	rom := &db.ROM{Name: "Super Mario Land (World).gb"}
	return match.Result{Path: path, Game: game, ROM: rom, Tier: match.TierSHA1}
}

func TestBuildFilePlan_MinUIUsesJapaneseTitle(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "old.gb")
	if err := os.WriteFile(src, []byte("rom"), 0o644); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildFilePlan(FileOptions{
		ROMDir:   root,
		Platform: "gb",
		Profile:  Profiles[FrontendMinUI],
		Format:   FileFormatRaw,
	}, []match.Result{matchedFile(src)})
	if err != nil {
		t.Fatal(err)
	}
	if got := filepath.Base(plan.Actions[0].Target); got != "スーパーマリオランド.gb" {
		t.Errorf("target: want Japanese title, got %q", got)
	}
}

func TestBuildFilePlan_ESDEUsesNoIntroName(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "old.gb")
	if err := os.WriteFile(src, []byte("rom"), 0o644); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildFilePlan(FileOptions{
		ROMDir:   root,
		Platform: "gb",
		Profile:  Profiles[FrontendESDE],
		Format:   FileFormatRaw,
	}, []match.Result{matchedFile(src)})
	if err != nil {
		t.Fatal(err)
	}
	if got := filepath.Base(plan.Actions[0].Target); got != "Super Mario Land (World).gb" {
		t.Errorf("target: want No-Intro name, got %q", got)
	}
}

func TestBuildFilePlan_MinUIDisambiguatesDuplicateJapaneseTitle(t *testing.T) {
	root := t.TempDir()
	src1 := filepath.Join(root, "Gallop Racer (Japan).chd")
	src2 := filepath.Join(root, "Gallop Racer 3 (Japan).chd")
	for _, src := range []string{src1, src2} {
		if err := os.WriteFile(src, []byte("rom"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	game := &db.Game{
		ID:     "gallop-racer",
		Titles: []db.Title{{Text: "ギャロップレーサー", Lang: "ja", Script: "Jpan", Verified: true}},
	}
	plan, err := BuildFilePlan(FileOptions{
		ROMDir:   root,
		Platform: "ps1",
		Profile:  Profiles[FrontendUnuOS],
		Format:   FileFormatRaw,
	}, []match.Result{
		{Path: src1, Game: game, ROM: &db.ROM{Name: "Gallop Racer (Japan).chd"}, Tier: match.TierSHA1},
		{Path: src2, Game: game, ROM: &db.ROM{Name: "Gallop Racer 3 (Japan).chd"}, Tier: match.TierSHA1},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range plan.Actions {
		if a.Status != StatusRename {
			t.Fatalf("%s: want rename, got %s (%s)", filepath.Base(a.Source), a.Status, a.Reason)
		}
	}
	got := map[string]bool{}
	for _, a := range plan.Actions {
		got[filepath.Base(a.Target)] = true
	}
	for _, want := range []string{
		"ギャロップレーサー (Gallop Racer (Japan)).chd",
		"ギャロップレーサー (Gallop Racer 3 (Japan)).chd",
	} {
		if !got[want] {
			t.Fatalf("missing target %q in %#v", want, got)
		}
	}
}

func TestApplyFilePlan_ZipRawROM(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "old.gb")
	if err := os.WriteFile(src, []byte("rom"), 0o644); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildFilePlan(FileOptions{
		ROMDir:   root,
		Platform: "gb",
		Profile:  Profiles[FrontendESDE],
		Format:   FileFormatZip,
	}, []match.Result{matchedFile(src)})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Actions[0].Operation != FileOpZip {
		t.Fatalf("operation: want zip, got %s", plan.Actions[0].Operation)
	}
	if _, err := ApplyFilePlan(plan, false, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("source should be removed after zip, stat err=%v", err)
	}
	target := filepath.Join(root, "Super Mario Land (World).zip")
	zr, err := zip.OpenReader(target)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	if len(zr.File) != 1 || zr.File[0].Name != "Super Mario Land (World).gb" {
		t.Fatalf("zip entry: got %#v", zr.File)
	}
}

func TestApplyFilePlan_UnzipROM(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "old.zip")
	if err := writeTestZip(src, "old.gb", []byte("rom")); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildFilePlan(FileOptions{
		ROMDir:   root,
		Platform: "gb",
		Profile:  Profiles[FrontendESDE],
		Format:   FileFormatRaw,
	}, []match.Result{matchedFile(src)})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Actions[0].Operation != FileOpUnzip {
		t.Fatalf("operation: want unzip, got %s", plan.Actions[0].Operation)
	}
	if _, err := ApplyFilePlan(plan, false, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("source zip should be removed after unzip, stat err=%v", err)
	}
	if b, err := os.ReadFile(filepath.Join(root, "Super Mario Land (World).gb")); err != nil || string(b) != "rom" {
		t.Fatalf("unzipped file: bytes=%q err=%v", string(b), err)
	}
}

func TestApplyFilePlan_RewritesZipInnerName(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "スーパーマリオランド.zip")
	if err := writeTestZip(src, "old.gb", []byte("rom")); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildFilePlan(FileOptions{
		ROMDir:   root,
		Platform: "gb",
		Profile:  Profiles[FrontendMinUI],
		Format:   FileFormatZip,
	}, []match.Result{matchedFile(src)})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Actions[0].Operation != FileOpRename {
		t.Fatalf("operation: want rename, got %s", plan.Actions[0].Operation)
	}
	if plan.Actions[0].Status != StatusRename {
		t.Fatalf("status: want rename, got %s", plan.Actions[0].Status)
	}
	if _, err := ApplyFilePlan(plan, false, nil); err != nil {
		t.Fatal(err)
	}

	zr, err := zip.OpenReader(src)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	if len(zr.File) != 1 || zr.File[0].Name != "スーパーマリオランド.gb" {
		t.Fatalf("zip entry: got %#v", zr.File)
	}
}

func writeTestZip(path, name string, data []byte) error {
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
