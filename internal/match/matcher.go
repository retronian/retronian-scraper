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
	Tier   Tier
}

type Matcher struct {
	games    []db.Game
	sha1Idx  map[string]*db.Game
	crc32Idx map[string]*db.Game
	md5Idx   map[string]*db.Game
}

func New(games []db.Game) *Matcher {
	m := &Matcher{
		games:    games,
		sha1Idx:  make(map[string]*db.Game, len(games)*2),
		crc32Idx: make(map[string]*db.Game, len(games)*2),
		md5Idx:   make(map[string]*db.Game, len(games)*2),
	}
	for i := range games {
		g := &games[i]
		for _, r := range g.ROMs {
			if v := strings.ToLower(r.SHA1); v != "" {
				m.sha1Idx[v] = g
			}
			if v := strings.ToLower(r.CRC32); v != "" {
				m.crc32Idx[v] = g
			}
			if v := strings.ToLower(r.MD5); v != "" {
				m.md5Idx[v] = g
			}
		}
	}
	return m
}

func (m *Matcher) Match(path string, h scan.Hashes) Result {
	r := Result{Path: path, Hashes: h}
	if g, ok := m.sha1Idx[strings.ToLower(h.SHA1)]; ok {
		r.Game = g
		r.Tier = TierSHA1
		return r
	}
	if g, ok := m.crc32Idx[strings.ToLower(h.CRC32)]; ok {
		r.Game = g
		r.Tier = TierHashFallback
		return r
	}
	if g, ok := m.md5Idx[strings.ToLower(h.MD5)]; ok {
		r.Game = g
		r.Tier = TierHashFallback
		return r
	}
	return r
}
