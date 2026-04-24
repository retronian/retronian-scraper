package pipeline

import (
	"os"
	"path/filepath"

	"github.com/retronian/retronian-scraper/internal/export"
)

// WriteGameList writes ES-DE / EmulationStation gamelist.xml for the given pipeline output.
func WriteGameList(out *Output, path, imageDir string) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return export.WriteESDE(f, out.Results, imageDir)
}
