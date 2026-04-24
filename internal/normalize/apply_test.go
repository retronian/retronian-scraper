package normalize

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func dirEntries(t *testing.T, root string) []string {
	t.Helper()
	es, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]string, 0, len(es))
	for _, e := range es {
		out = append(out, e.Name())
	}
	sort.Strings(out)
	return out
}

func TestApply_DryRun_NoFilesystemChange(t *testing.T) {
	root := t.TempDir()
	mkSubdir(t, root, "GameBoy")

	plan, err := BuildPlan(root, Profiles[FrontendESDE])
	if err != nil {
		t.Fatal(err)
	}
	res, err := Apply(plan, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Renamed) != 1 {
		t.Errorf("dry-run: want 1 renamed (logical), got %d", len(res.Renamed))
	}
	got := dirEntries(t, root)
	want := []string{"GameBoy"}
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("filesystem changed during dry-run: got %v, want %v", got, want)
	}
}

func TestApply_RealRename(t *testing.T) {
	root := t.TempDir()
	mkSubdir(t, root, "GameBoy")
	mkSubdir(t, root, "snes")
	mkSubdir(t, root, "Misc")
	mkFile(t, filepath.Join(root, "Misc"), "readme.txt")

	plan, err := BuildPlan(root, Profiles[FrontendESDE])
	if err != nil {
		t.Fatal(err)
	}
	res, err := Apply(plan, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Renamed) != 1 {
		t.Errorf("renamed: want 1, got %d", len(res.Renamed))
	}
	if len(res.NoOp) != 1 {
		t.Errorf("noop: want 1, got %d", len(res.NoOp))
	}
	if len(res.Unknown) != 1 {
		t.Errorf("unknown: want 1, got %d", len(res.Unknown))
	}
	if len(res.Errors) != 0 {
		t.Errorf("errors: want 0, got %d", len(res.Errors))
	}

	got := dirEntries(t, root)
	want := []string{"Misc", "gb", "snes"}
	if !equalStringSlices(got, want) {
		t.Errorf("filesystem after apply: got %v, want %v", got, want)
	}
}

func TestApply_OnActionCallback(t *testing.T) {
	root := t.TempDir()
	mkSubdir(t, root, "GameBoy")
	mkSubdir(t, root, "snes")

	plan, err := BuildPlan(root, Profiles[FrontendESDE])
	if err != nil {
		t.Fatal(err)
	}
	var seen []ActionStatus
	_, err = Apply(plan, true, func(a Action) {
		seen = append(seen, a.Status)
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != 2 {
		t.Errorf("callback fired %d times, want 2", len(seen))
	}
}

func TestApply_ConflictAndUnknownNotRenamed(t *testing.T) {
	root := t.TempDir()
	mkSubdir(t, root, "GameBoy")
	mkSubdir(t, root, "gb") // pre-existing target → conflict for GameBoy
	misc := mkSubdir(t, root, "Misc")
	mkFile(t, misc, "x.txt")

	plan, err := BuildPlan(root, Profiles[FrontendESDE])
	if err != nil {
		t.Fatal(err)
	}
	res, err := Apply(plan, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Conflicts) != 1 || len(res.Unknown) != 1 {
		t.Errorf("conflicts=%d unknown=%d, want 1/1", len(res.Conflicts), len(res.Unknown))
	}
	got := dirEntries(t, root)
	want := []string{"GameBoy", "Misc", "gb"} // unchanged
	if !equalStringSlices(got, want) {
		t.Errorf("filesystem changed despite conflict: got %v, want %v", got, want)
	}
}

func TestRenameDir_CaseOnly(t *testing.T) {
	// On case-insensitive FS this exercises the 2-step path. On
	// case-sensitive FS it just behaves like a normal rename.
	root := t.TempDir()
	src := filepath.Join(root, "Foo")
	if err := os.Mkdir(src, 0o755); err != nil {
		t.Fatal(err)
	}
	tgt := filepath.Join(root, "foo")
	if err := renameDir(src, tgt); err != nil {
		t.Fatalf("renameDir: %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d (%v)", len(entries), entries)
	}
	if entries[0].Name() != "foo" {
		t.Errorf("rename to %q failed: got %q", "foo", entries[0].Name())
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
