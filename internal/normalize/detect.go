package normalize

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/retronian/retronian-scraper/internal/scan"
)

type DetectionMethod string

const (
	DetectByAlias    DetectionMethod = "alias"
	DetectByContents DetectionMethod = "contents"
	DetectFailed     DetectionMethod = "failed"
)

// Detection records how a platform was identified for one input folder.
type Detection struct {
	InternalID string
	Method     DetectionMethod
	Score      int    // contents: hit count for the winning platform; alias: 0
	Note       string // human-readable explanation
}

// tieBreakOrder defines the priority used to resolve content-detection
// ties. Earlier entries win. Chosen so platforms with mostly-unambiguous
// extensions float to the top, leaving ambiguous .bin to the broader
// systems (md / ps1).
var tieBreakOrder = []string{
	"n64", "nds", "ps1", "sfc", "fc", "md", "pce", "gba", "gbc", "gb",
}

var tieBreakRank = func() map[string]int {
	out := make(map[string]int, len(tieBreakOrder))
	for i, p := range tieBreakOrder {
		out[p] = i
	}
	return out
}()

// extToPlatforms maps a lower-case extension (with leading dot) to the
// list of platform IDs that recognise it. An extension may belong to
// multiple platforms (e.g. .bin → md, ps1).
var extToPlatforms = func() map[string][]string {
	out := make(map[string][]string, 32)
	for _, id := range scan.KnownPlatforms() {
		for _, e := range scan.PlatformExtensions(id) {
			out[e] = append(out[e], id)
		}
	}
	return out
}()

// DetectPlatform identifies the internal platform ID for a single
// subfolder. dir is the absolute path of the folder; name is its
// basename used for alias lookup.
//
// Strategy:
//  1. Try alias lookup against the folder name.
//  2. On miss, walk dir recursively and count how many files map to each
//     platform via extToPlatforms. The winner is the platform with the
//     highest count; ties are broken by tieBreakOrder.
//  3. If neither alias nor contents identify a platform, return DetectFailed.
func DetectPlatform(dir, name string) (Detection, error) {
	if id, ok := LookupByAlias(NormalizeFolderName(name)); ok {
		return Detection{
			InternalID: id,
			Method:     DetectByAlias,
			Note:       fmt.Sprintf("matched alias from folder name %q", name),
		}, nil
	}

	counts, total, err := countByExtension(dir)
	if err != nil {
		return Detection{}, err
	}
	if total == 0 {
		return Detection{
			Method: DetectFailed,
			Note:   "no alias match and no recognised ROM extensions in folder",
		}, nil
	}

	winner, score := pickWinner(counts)
	if winner == "" {
		return Detection{
			Method: DetectFailed,
			Note:   "no alias match and no recognised ROM extensions in folder",
		}, nil
	}
	return Detection{
		InternalID: winner,
		Method:     DetectByContents,
		Score:      score,
		Note: fmt.Sprintf("inferred from %d/%d ROM file(s) in folder",
			score, total),
	}, nil
}

func countByExtension(dir string) (map[string]int, int, error) {
	counts := map[string]int{}
	total := 0
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		platforms, ok := extToPlatforms[ext]
		if !ok {
			return nil
		}
		total++
		for _, p := range platforms {
			counts[p]++
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	return counts, total, nil
}

func pickWinner(counts map[string]int) (string, int) {
	var (
		best     string
		bestN    int
		bestRank = len(tieBreakOrder) + 1
	)
	for p, n := range counts {
		if n == 0 {
			continue
		}
		rank, ok := tieBreakRank[p]
		if !ok {
			rank = len(tieBreakOrder)
		}
		switch {
		case n > bestN:
			best, bestN, bestRank = p, n, rank
		case n == bestN && rank < bestRank:
			best, bestRank = p, rank
		}
	}
	return best, bestN
}
