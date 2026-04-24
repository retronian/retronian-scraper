package scan

import (
	"io/fs"
	"path/filepath"
	"strings"
)

var platformExts = map[string][]string{
	"fc":  {".nes", ".fds"},
	"sfc": {".sfc", ".smc"},
	"gb":  {".gb"},
	"gbc": {".gbc"},
	"gba": {".gba"},
	"md":  {".md", ".gen", ".smd", ".bin"},
	"pce": {".pce"},
	"n64": {".n64", ".z64", ".v64"},
	"nds": {".nds"},
	"ps1": {".iso", ".bin", ".cue", ".img", ".chd"},
}

func Walk(root, platform string) ([]string, error) {
	exts, ok := platformExts[platform]
	if !ok {
		return nil, &UnknownPlatformError{Platform: platform}
	}
	extSet := make(map[string]struct{}, len(exts))
	for _, e := range exts {
		extSet[e] = struct{}{}
	}

	var out []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if _, ok := extSet[strings.ToLower(filepath.Ext(path))]; ok {
			out = append(out, path)
		}
		return nil
	})
	return out, err
}

type UnknownPlatformError struct{ Platform string }

func (e *UnknownPlatformError) Error() string {
	return "unknown platform: " + e.Platform
}

func KnownPlatforms() []string {
	out := make([]string, 0, len(platformExts))
	for p := range platformExts {
		out = append(out, p)
	}
	return out
}
