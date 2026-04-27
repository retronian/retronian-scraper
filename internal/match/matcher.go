package match

import (
	"archive/zip"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/retronian/retronian-scraper/internal/db"
	"github.com/retronian/retronian-scraper/internal/scan"
)

type Tier int

const (
	TierNone Tier = iota
	TierSHA1
	TierSlug
	TierHashFallback
	TierNameFallback
)

type Result struct {
	Path   string
	Hashes scan.Hashes
	Game   *db.Game
	ROM    *db.ROM
	Tier   Tier
}

type matchedROM struct {
	game *db.Game
	rom  *db.ROM
}

type Matcher struct {
	games    []db.Game
	sha1Idx  map[string]matchedROM
	crc32Idx map[string]matchedROM
	md5Idx   map[string]matchedROM
	nameIdx  map[string]matchedROM
}

var separatorRE = regexp.MustCompile(`\s+`)

func New(games []db.Game) *Matcher {
	m := &Matcher{
		games:    games,
		sha1Idx:  make(map[string]matchedROM, len(games)*2),
		crc32Idx: make(map[string]matchedROM, len(games)*2),
		md5Idx:   make(map[string]matchedROM, len(games)*2),
		nameIdx:  make(map[string]matchedROM, len(games)*2),
	}
	for i := range games {
		g := &games[i]
		for j := range g.ROMs {
			r := &g.ROMs[j]
			match := matchedROM{game: g, rom: r}
			if v := strings.ToLower(r.SHA1); v != "" {
				m.sha1Idx[v] = match
			}
			if v := strings.ToLower(r.CRC32); v != "" {
				m.crc32Idx[v] = match
			}
			if v := strings.ToLower(r.MD5); v != "" {
				m.md5Idx[v] = match
			}
			if v := canonicalROMName(r.Name); v != "" {
				m.nameIdx[v] = match
			}
		}
		for j := range g.Titles {
			if v := canonicalROMName(g.Titles[j].Text); v != "" {
				m.nameIdx[v] = matchedROM{game: g}
			}
		}
	}
	return m
}

func (m *Matcher) Match(path string, h scan.Hashes) Result {
	r := Result{Path: path, Hashes: h}
	if match, ok := m.sha1Idx[strings.ToLower(h.SHA1)]; ok {
		r.Game = match.game
		r.ROM = match.rom
		r.Tier = TierSHA1
		return r
	}
	if match, ok := m.crc32Idx[strings.ToLower(h.CRC32)]; ok {
		r.Game = match.game
		r.ROM = match.rom
		r.Tier = TierHashFallback
		return r
	}
	if match, ok := m.md5Idx[strings.ToLower(h.MD5)]; ok {
		r.Game = match.game
		r.ROM = match.rom
		r.Tier = TierHashFallback
		return r
	}
	if match, ok := m.nameIdx[canonicalROMName(matchName(path))]; ok {
		r.Game = match.game
		r.ROM = match.rom
		r.Tier = TierNameFallback
		return r
	}
	return r
}

func matchName(path string) string {
	if strings.EqualFold(filepath.Ext(path), ".zip") {
		return firstZipEntryName(path)
	}
	return filepath.Base(path)
}

func firstZipEntryName(path string) string {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return ""
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.FileInfo().IsDir() || strings.HasPrefix(filepath.Base(f.Name), ".") {
			continue
		}
		return filepath.Base(f.Name)
	}
	return ""
}

func canonicalROMName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	for {
		ext := strings.ToLower(filepath.Ext(name))
		switch ext {
		case ".zip", ".7z", ".nes", ".fds", ".sfc", ".smc", ".gb", ".gbc", ".gba", ".md", ".gen", ".smd", ".pce", ".chd", ".cue", ".bin", ".iso":
			name = strings.TrimSpace(strings.TrimSuffix(name, filepath.Ext(name)))
		default:
			name = stripTrailingRegionTags(name)
			name = strings.NewReplacer(
				"：", " ",
				":", " ",
				" - ", " ",
				" – ", " ",
				" — ", " ",
			).Replace(name)
			name = separatorRE.ReplaceAllString(name, " ")
			return strings.ToLower(strings.TrimSpace(name))
		}
	}
}

func stripTrailingRegionTags(name string) string {
	for {
		trimmed := strings.TrimSpace(name)
		open := strings.LastIndex(trimmed, "(")
		if open == -1 || !strings.HasSuffix(trimmed, ")") {
			return trimmed
		}
		tag := strings.ToLower(strings.TrimSpace(trimmed[open+1 : len(trimmed)-1]))
		if !isTrailingRegionTag(tag) {
			return trimmed
		}
		name = strings.TrimSpace(trimmed[:open])
	}
}

func isTrailingRegionTag(tag string) bool {
	switch tag {
	case "japan", "usa", "europe", "world", "en", "ja":
		return true
	default:
		return strings.HasPrefix(tag, "rev ") || strings.HasPrefix(tag, "disc ")
	}
}
