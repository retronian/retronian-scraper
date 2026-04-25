package scan

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestHash_ZipHashesInnerROM(t *testing.T) {
	root := t.TempDir()
	raw := filepath.Join(root, "game.gb")
	data := []byte("rom")
	if err := os.WriteFile(raw, data, 0o644); err != nil {
		t.Fatal(err)
	}
	want, err := Hash(raw)
	if err != nil {
		t.Fatal(err)
	}

	zpath := filepath.Join(root, "game.zip")
	out, err := os.Create(zpath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(out)
	w, err := zw.Create("game.gb")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := Hash(zpath)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("zip hash: got %+v, want %+v", got, want)
	}
}
