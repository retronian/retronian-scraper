package normalize

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func mkSubdir(t *testing.T, root, name string) string {
	t.Helper()
	p := filepath.Join(root, name)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func mkFile(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func findAction(t *testing.T, plan *Plan, basename string) Action {
	t.Helper()
	for _, a := range plan.Actions {
		if filepath.Base(a.Source) == basename {
			return a
		}
	}
	t.Fatalf("no action for %q in plan", basename)
	return Action{}
}

func TestBuildPlan_RenameAndNoop(t *testing.T) {
	root := t.TempDir()
	mkSubdir(t, root, "GameBoy")
	mkSubdir(t, root, "snes")

	plan, err := BuildPlan(root, Profiles[FrontendESDE])
	if err != nil {
		t.Fatal(err)
	}
	if got := len(plan.Actions); got != 2 {
		t.Fatalf("len actions: want 2, got %d", got)
	}

	gb := findAction(t, plan, "GameBoy")
	if gb.Status != StatusRename || filepath.Base(gb.Target) != "gb" {
		t.Errorf("GameBoy: want rename → gb, got %s → %s", gb.Status, filepath.Base(gb.Target))
	}

	snes := findAction(t, plan, "snes")
	if snes.Status != StatusNoop {
		t.Errorf("snes: want noop, got %s", snes.Status)
	}
}

func TestBuildPlanForLanguage_JapaneseFolderNames(t *testing.T) {
	root := t.TempDir()
	mkSubdir(t, root, "GameBoy")
	mkSubdir(t, root, "スーパーファミコン (SFC)")

	plan, err := BuildPlanForLanguage(root, Profiles[FrontendMinUI], LanguageJapanese)
	if err != nil {
		t.Fatal(err)
	}

	gb := findAction(t, plan, "GameBoy")
	if gb.Status != StatusRename || filepath.Base(gb.Target) != "ゲームボーイ (GB)" {
		t.Errorf("GameBoy: want rename → ゲームボーイ (GB), got %s → %s", gb.Status, filepath.Base(gb.Target))
	}

	sfc := findAction(t, plan, "スーパーファミコン (SFC)")
	if sfc.Status != StatusNoop {
		t.Errorf("スーパーファミコン (SFC): want noop, got %s", sfc.Status)
	}
}

func TestBuildPlanForLanguage_AllSupportedLanguages(t *testing.T) {
	wantGB := map[LanguageID]string{
		LanguageEnglish:  "Game Boy (GB)",
		LanguageJapanese: "ゲームボーイ (GB)",
		LanguageKorean:   "게임보이 (GB)",
		LanguageChinese:  "Game Boy (GB)",
		LanguageFrench:   "Game Boy (GB)",
		LanguageSpanish:  "Game Boy (GB)",
		LanguageGerman:   "Game Boy (GB)",
	}
	for _, langText := range KnownLanguages() {
		lang := LanguageID(langText)
		root := t.TempDir()
		mkSubdir(t, root, "gb")

		plan, err := BuildPlanForLanguage(root, Profiles[FrontendMinUI], lang)
		if err != nil {
			t.Fatal(err)
		}
		a := findAction(t, plan, "gb")
		if filepath.Base(a.Target) != wantGB[lang] {
			t.Errorf("lang %s: want %q, got %q", lang, wantGB[lang], filepath.Base(a.Target))
		}
	}
}

func TestBuildPlan_HiddenIgnored(t *testing.T) {
	root := t.TempDir()
	mkSubdir(t, root, ".git")
	mkSubdir(t, root, ".cache")
	mkSubdir(t, root, "gb")

	plan, err := BuildPlan(root, Profiles[FrontendESDE])
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range plan.Actions {
		base := filepath.Base(a.Source)
		if base == ".git" || base == ".cache" {
			t.Errorf("hidden folder leaked into plan: %s", base)
		}
	}
}

func TestBuildPlan_Symlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks not exercised on windows")
	}
	root := t.TempDir()
	target := mkSubdir(t, root, "actual")
	link := filepath.Join(root, "linked")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildPlan(root, Profiles[FrontendESDE])
	if err != nil {
		t.Fatal(err)
	}
	a := findAction(t, plan, "linked")
	if a.Status != StatusSkipped || a.Reason != "symbolic link" {
		t.Errorf("symlink action: want skipped/symbolic link, got %s/%s", a.Status, a.Reason)
	}
}

func TestBuildPlan_Unknown(t *testing.T) {
	root := t.TempDir()
	misc := mkSubdir(t, root, "Misc")
	mkFile(t, misc, "readme.txt")

	plan, err := BuildPlan(root, Profiles[FrontendESDE])
	if err != nil {
		t.Fatal(err)
	}
	a := findAction(t, plan, "Misc")
	if a.Status != StatusUnknown {
		t.Errorf("Misc: want unknown, got %s", a.Status)
	}
}

func TestBuildPlan_TargetExistsConflict(t *testing.T) {
	root := t.TempDir()
	mkSubdir(t, root, "GameBoy")
	mkSubdir(t, root, "gb") // already-normalised folder, distinct from GameBoy

	plan, err := BuildPlan(root, Profiles[FrontendESDE])
	if err != nil {
		t.Fatal(err)
	}
	gb := findAction(t, plan, "GameBoy")
	if gb.Status != StatusConflict {
		t.Errorf("GameBoy: want conflict, got %s (reason: %s)", gb.Status, gb.Reason)
	}
}

func TestBuildPlan_MultipleSourcesSameTarget(t *testing.T) {
	root := t.TempDir()
	mkSubdir(t, root, "GameBoy")
	mkSubdir(t, root, "Game-Boy")

	plan, err := BuildPlan(root, Profiles[FrontendESDE])
	if err != nil {
		t.Fatal(err)
	}
	a1 := findAction(t, plan, "GameBoy")
	a2 := findAction(t, plan, "Game-Boy")
	if a1.Status != StatusConflict || a2.Status != StatusConflict {
		t.Errorf("both should conflict, got %s and %s", a1.Status, a2.Status)
	}
}

func TestBuildPlan_FallbackPlatform(t *testing.T) {
	root := t.TempDir()
	mkSubdir(t, root, "n64")

	// minui has no n64 mapping → fallback to "n64" → noop.
	plan, err := BuildPlan(root, Profiles[FrontendMinUI])
	if err != nil {
		t.Fatal(err)
	}
	a := findAction(t, plan, "n64")
	if !a.Fallback {
		t.Errorf("expected Fallback=true for minui n64")
	}
	if a.Status != StatusNoop {
		t.Errorf("minui n64 noop: got status %s", a.Status)
	}
}

func TestBuildPlan_BasenameSorted(t *testing.T) {
	root := t.TempDir()
	mkSubdir(t, root, "zoo")
	mkSubdir(t, root, "alpha")
	mkSubdir(t, root, "mid")

	plan, err := BuildPlan(root, Profiles[FrontendESDE])
	if err != nil {
		t.Fatal(err)
	}
	prev := ""
	for _, a := range plan.Actions {
		base := filepath.Base(a.Source)
		if prev != "" && base < prev {
			t.Errorf("not sorted: %s before %s", prev, base)
		}
		prev = base
	}
}

func TestBuildPlan_RootMissing(t *testing.T) {
	if _, err := BuildPlan("/nonexistent/path/x", Profiles[FrontendESDE]); err == nil {
		t.Errorf("expected error for missing root")
	}
}
