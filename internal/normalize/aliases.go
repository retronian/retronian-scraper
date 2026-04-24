package normalize

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

// aliasTable maps internal platform IDs to a list of canonical aliases.
// An alias is a normalised key produced by NormalizeFolderName, used to
// recognise common ROM folder names across the supported frontends.
//
// Each entry includes both short codes (gb, sfc) and the official folder
// names of all six supported frontends so that re-normalising an
// already-organised collection from one frontend to another succeeds.
var aliasTable = map[string][]string{
	"fc": {
		"fc", "famicom", "nes",
		"nintendoentertainmentsystem",
		"nintendoentertainmentsystemfc",
	},
	"sfc": {
		"sfc", "snes", "superfamicom", "supernintendo",
		"supernintendoentertainmentsystem",
		"supernintendoentertainmentsystemsfc",
	},
	"gb": {
		"gb", "gameboy", "gameboygb",
	},
	"gbc": {
		"gbc", "gameboycolor", "gameboycolorgbc",
	},
	"gba": {
		"gba", "gameboyadvance", "gameboyadvancegba",
	},
	"md": {
		"md", "megadrive", "genesis", "segagenesis", "segagenesismd",
	},
	"pce": {
		"pce", "pcengine", "turbografx", "turbografx16", "turbografx16pce",
	},
	"n64": {
		"n64", "nintendo64",
	},
	"nds": {
		"nds", "nintendods", "ds",
	},
	"ps1": {
		"ps1", "ps", "psx", "playstation",
		"sonyplaystation", "sonyplaystationps",
	},
}

// aliasIndex is the reverse lookup canonical alias → internal ID.
// Built once at init time.
var aliasIndex = func() map[string]string {
	out := make(map[string]string, 64)
	for id, aliases := range aliasTable {
		for _, a := range aliases {
			if existing, dup := out[a]; dup && existing != id {
				panic("normalize: duplicate alias " + a +
					" maps to both " + existing + " and " + id)
			}
			out[a] = id
		}
	}
	return out
}()

// NormalizeFolderName produces a canonical key from an arbitrary folder
// name by applying NFKC, lower-casing, and stripping common separator
// punctuation. Returns "" for inputs that contain no alphanumerics.
func NormalizeFolderName(s string) string {
	s = norm.NFKC.String(s)
	s = strings.ToLower(s)

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case ' ', '-', '_', '.', '(', ')', '[', ']', '/', '\\', '\t':
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// LookupByAlias returns the internal platform ID for an already
// canonicalised alias key. Pass NormalizeFolderName output here.
func LookupByAlias(canonical string) (string, bool) {
	if canonical == "" {
		return "", false
	}
	id, ok := aliasIndex[canonical]
	return id, ok
}
