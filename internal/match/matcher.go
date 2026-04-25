package match

import (
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
}

func New(games []db.Game) *Matcher {
	m := &Matcher{
		games:    games,
		sha1Idx:  make(map[string]matchedROM, len(games)*2),
		crc32Idx: make(map[string]matchedROM, len(games)*2),
		md5Idx:   make(map[string]matchedROM, len(games)*2),
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
	return r
}
