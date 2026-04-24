package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Note: Go's standard flag package stops parsing at the first non-flag
// argument, so positional args must come AFTER flags. `scan` follows the
// same convention.

func TestNormalize_MissingFrontend(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runNormalize([]string{t.TempDir()}, &stdout, &stderr)
	if code != 2 {
		t.Errorf("missing --frontend: want exit 2, got %d", code)
	}
}

func TestNormalize_UnknownFrontend(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runNormalize([]string{"--frontend", "retroarch", t.TempDir()}, &stdout, &stderr)
	if code != 2 {
		t.Errorf("unknown frontend: want exit 2, got %d", code)
	}
}

func TestNormalize_MissingDir(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runNormalize([]string{"--frontend", "es-de", "/nonexistent/x"}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("missing dir: want exit 1, got %d", code)
	}
}

func TestNormalize_DryRunWithRename(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "GameBoy"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := runNormalize([]string{"--frontend", "es-de", root}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("dry-run no conflict: want exit 0, got %d (stderr: %s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "rename") {
		t.Errorf("output missing 'rename' tag:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "re-run with --apply") {
		t.Errorf("output missing 're-run with --apply':\n%s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(root, "GameBoy")); err != nil {
		t.Errorf("dry-run modified filesystem: %v", err)
	}
}

func TestNormalize_ConflictExitsOne(t *testing.T) {
	root := t.TempDir()
	for _, d := range []string{"GameBoy", "gb"} {
		if err := os.Mkdir(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	var stdout, stderr bytes.Buffer
	code := runNormalize([]string{"--frontend", "es-de", root}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("conflict present: want exit 1, got %d", code)
	}
	if !strings.Contains(stdout.String(), "conflict") {
		t.Errorf("output missing 'conflict':\n%s", stdout.String())
	}
}

func TestNormalize_ApplyPerformsRename(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "GameBoy"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := runNormalize([]string{"--frontend", "es-de", "--apply", root}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("apply no conflict: want exit 0, got %d (stderr: %s)", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "gb")); err != nil {
		t.Errorf("apply did not rename: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "GameBoy")); err == nil {
		t.Errorf("apply left source folder behind")
	}
}
