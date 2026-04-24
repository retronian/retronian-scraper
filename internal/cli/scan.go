package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/retronian/retronian-scraper/internal/db"
	"github.com/retronian/retronian-scraper/internal/export"
	"github.com/retronian/retronian-scraper/internal/match"
	"github.com/retronian/retronian-scraper/internal/scan"
)

func Scan(args []string) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	platform := fs.String("platform", "", "platform id (gb, gbc, gba, fc, sfc, md, pce, n64, nds, ps1)")
	out := fs.String("out", "gamelist.xml", "output gamelist.xml path")
	imageDir := fs.String("image-dir", "./images/", "relative image directory referenced in gamelist.xml")
	baseURL := fs.String("api", db.DefaultBaseURL, "native-game-db API base URL")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: retronian-scraper scan <rom-dir> --platform <id> [--out gamelist.xml]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() < 1 || *platform == "" {
		fs.Usage()
		return 2
	}
	romDir := fs.Arg(0)

	paths, err := scan.Walk(romDir, *platform)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if len(paths) == 0 {
		fmt.Fprintf(os.Stderr, "no ROMs found under %s for platform %s\n", romDir, *platform)
		return 1
	}
	fmt.Printf("found %d ROM(s)\n", len(paths))

	hashes, err := hashAll(context.Background(), paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	apiURL := fmt.Sprintf("%s/api/v1/%s.json", *baseURL, *platform)
	fmt.Printf("fetching DB: %s\n", apiURL)
	client := db.NewClient()
	client.BaseURL = *baseURL
	games, err := client.PlatformGames(*platform)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("%d games in DB\n", len(games))

	matcher := match.New(games)
	results := make([]match.Result, len(paths))
	tierCount := map[match.Tier]int{}
	for i, p := range paths {
		results[i] = matcher.Match(p, hashes[i])
		tierCount[results[i].Tier]++
	}

	matched := tierCount[match.TierSHA1] + tierCount[match.TierSlug] + tierCount[match.TierHashFallback]
	fmt.Printf("matched %d/%d (sha1=%d, fallback_hash=%d, unmatched=%d)\n",
		matched, len(results),
		tierCount[match.TierSHA1],
		tierCount[match.TierHashFallback],
		tierCount[match.TierNone],
	)

	if dir := filepath.Dir(*out); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	f, err := os.Create(*out)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer f.Close()
	if err := export.WriteESDE(f, results, *imageDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("wrote %s\n", *out)
	return 0
}

func hashAll(ctx context.Context, paths []string) ([]scan.Hashes, error) {
	hashes := make([]scan.Hashes, len(paths))
	errs := make([]error, len(paths))
	var done int64
	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU())

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
			if int(n) == len(paths) {
				fmt.Fprintf(os.Stderr, "\rhashing %d/%d\n", n, len(paths))
			} else if int(n)%50 == 0 {
				fmt.Fprintf(os.Stderr, "\rhashing %d/%d", n, len(paths))
			}
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
