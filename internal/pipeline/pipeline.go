// Package pipeline orchestrates the scan → hash → fetch → match flow used by
// both the CLI and the GUI. Progress is reported through a callback so that
// callers can render it however they prefer (stderr, progress bar, etc.).
package pipeline

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/retronian/retronian-scraper/internal/db"
	"github.com/retronian/retronian-scraper/internal/match"
	"github.com/retronian/retronian-scraper/internal/scan"
)

type Options struct {
	ROMDir   string
	Platform string
	BaseURL  string
}

type Phase string

const (
	PhaseWalking  Phase = "walking"
	PhaseHashing  Phase = "hashing"
	PhaseFetching Phase = "fetching"
	PhaseMatching Phase = "matching"
	PhaseDone     Phase = "done"
)

type Progress struct {
	Phase Phase
	Done  int
	Total int
	Msg   string
}

type Output struct {
	Results   []match.Result
	TierCount map[match.Tier]int
}

func Run(ctx context.Context, opts Options, onProgress func(Progress)) (*Output, error) {
	report := func(p Progress) {
		if onProgress != nil {
			onProgress(p)
		}
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = db.DefaultBaseURL
	}

	report(Progress{Phase: PhaseWalking, Msg: opts.ROMDir})
	paths, err := scan.Walk(opts.ROMDir, opts.Platform)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no ROMs found under %s for platform %s", opts.ROMDir, opts.Platform)
	}
	report(Progress{
		Phase: PhaseWalking, Done: len(paths), Total: len(paths),
		Msg: fmt.Sprintf("found %d ROM(s)", len(paths)),
	})

	hashes, err := hashAll(ctx, paths, report)
	if err != nil {
		return nil, err
	}

	report(Progress{
		Phase: PhaseFetching,
		Msg:   fmt.Sprintf("%s/api/v1/%s.json", baseURL, opts.Platform),
	})
	client := db.NewClient()
	client.BaseURL = baseURL
	games, err := client.PlatformGames(opts.Platform)
	if err != nil {
		return nil, err
	}
	report(Progress{
		Phase: PhaseFetching, Done: len(games), Total: len(games),
		Msg: fmt.Sprintf("%d games in DB", len(games)),
	})

	matcher := match.New(games)
	results := make([]match.Result, len(paths))
	tierCount := map[match.Tier]int{}
	for i, p := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		results[i] = matcher.Match(p, hashes[i])
		tierCount[results[i].Tier]++
		report(Progress{Phase: PhaseMatching, Done: i + 1, Total: len(paths)})
	}

	out := &Output{Results: results, TierCount: tierCount}
	report(Progress{
		Phase: PhaseDone, Done: len(paths), Total: len(paths),
		Msg: summary(tierCount, len(paths)),
	})
	return out, nil
}

func hashAll(ctx context.Context, paths []string, report func(Progress)) ([]scan.Hashes, error) {
	hashes := make([]scan.Hashes, len(paths))
	errs := make([]error, len(paths))
	var done int64
	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU())

	report(Progress{Phase: PhaseHashing, Done: 0, Total: len(paths)})

	for i, p := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, p string) {
			defer wg.Done()
			defer func() { <-sem }()
			h, err := scan.Hash(p)
			if err != nil {
				errs[i] = fmt.Errorf("hash %s: %w", p, err)
				return
			}
			hashes[i] = h
			n := atomic.AddInt64(&done, 1)
			report(Progress{Phase: PhaseHashing, Done: int(n), Total: len(paths)})
		}(i, p)
	}
	wg.Wait()

	for _, e := range errs {
		if e != nil {
			return nil, e
		}
	}
	return hashes, nil
}

func summary(tier map[match.Tier]int, total int) string {
	matched := tier[match.TierSHA1] + tier[match.TierSlug] + tier[match.TierHashFallback]
	return fmt.Sprintf("matched %d/%d (sha1=%d, fallback_hash=%d, unmatched=%d)",
		matched, total,
		tier[match.TierSHA1],
		tier[match.TierHashFallback],
		tier[match.TierNone],
	)
}

// Summary formats an Output's tier counts the same way as the final progress message.
func Summary(out *Output) string {
	if out == nil {
		return ""
	}
	return summary(out.TierCount, len(out.Results))
}
