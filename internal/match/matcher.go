package match

import (
	"archive/zip"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/retronian/retronian-scraper/internal/db"
	"github.com/retronian/retronian-scraper/internal/scan"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/width"
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
	nameIdx  map[string]matchedROM
	crc32Idx map[string]matchedROM
	md5Idx   map[string]matchedROM
}

func New(games []db.Game) *Matcher {
	m := &Matcher{
		games:    games,
		sha1Idx:  make(map[string]matchedROM, len(games)*2),
		nameIdx:  make(map[string]matchedROM, len(games)*4),
		crc32Idx: make(map[string]matchedROM, len(games)*2),
		md5Idx:   make(map[string]matchedROM, len(games)*2),
	}
	ambiguousNames := map[string]struct{}{}
	for i := range games {
		g := &games[i]
		for j := range g.ROMs {
			r := &g.ROMs[j]
			match := matchedROM{game: g, rom: r}
			if v := strings.ToLower(r.SHA1); v != "" {
				m.sha1Idx[v] = match
			}
			addNameMatch(m.nameIdx, ambiguousNames, canonicalROMName(r.Name), match)
			if v := strings.ToLower(r.CRC32); v != "" {
				m.crc32Idx[v] = match
			}
			if v := strings.ToLower(r.MD5); v != "" {
				m.md5Idx[v] = match
			}
		}
		for j := range g.Titles {
			addNameMatch(m.nameIdx, ambiguousNames, canonicalROMName(g.Titles[j].Text), matchedROM{game: g})
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

func addNameMatch(idx map[string]matchedROM, ambiguous map[string]struct{}, key string, match matchedROM) {
	if key == "" {
		return
	}
	if existing, exists := idx[key]; exists {
		if existing.game != match.game {
			if len(existing.game.ROMs) == 0 && len(match.game.ROMs) > 0 {
				idx[key] = match
				return
			}
			if len(existing.game.ROMs) > 0 && len(match.game.ROMs) == 0 {
				return
			}
			delete(idx, key)
			ambiguous[key] = struct{}{}
		}
		return
	}
	if _, ok := ambiguous[key]; !ok {
		idx[key] = match
	}
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
		case ".zip", ".7z", ".nes", ".fds", ".sfc", ".smc", ".gb", ".gbc", ".gba", ".md", ".gen", ".smd", ".bin", ".pce", ".n64", ".z64", ".v64", ".nds", ".iso", ".cue", ".img", ".chd", ".pbp":
			name = strings.TrimSpace(strings.TrimSuffix(name, filepath.Ext(name)))
		default:
			name = stripTrailingRegionTags(name)
			name = strings.ToLower(width.Fold.String(norm.NFKC.String(name)))
			var b strings.Builder
			for _, r := range name {
				if unicode.IsLetter(r) || unicode.IsDigit(r) {
					b.WriteRune(r)
				}
			}
			return b.String()
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
