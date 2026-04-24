package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/retronian/retronian-scraper/internal/db"
	"github.com/retronian/retronian-scraper/internal/match"
	"github.com/retronian/retronian-scraper/internal/pipeline"
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

	output, err := pipeline.Run(context.Background(), pipeline.Options{
		ROMDir:   fs.Arg(0),
		Platform: *platform,
		BaseURL:  *baseURL,
	}, cliProgress())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	matched := output.TierCount[match.TierSHA1] + output.TierCount[match.TierSlug] + output.TierCount[match.TierHashFallback]
	fmt.Printf("matched %d/%d (sha1=%d, fallback_hash=%d, unmatched=%d)\n",
		matched, len(output.Results),
		output.TierCount[match.TierSHA1],
		output.TierCount[match.TierHashFallback],
		output.TierCount[match.TierNone],
	)

	if err := pipeline.WriteGameList(output, *out, *imageDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("wrote %s\n", *out)
	return 0
}

// cliProgress returns a progress callback that mirrors the prior CLI output:
// - walking complete:  "found N ROM(s)"
// - hashing:           "\rhashing N/M" every 50 items (and at completion)
// - fetching start:    "fetching DB: <url>"
// - fetching complete: "M games in DB"
func cliProgress() func(pipeline.Progress) {
	return func(p pipeline.Progress) {
		switch p.Phase {
		case pipeline.PhaseWalking:
			if p.Total > 0 {
				fmt.Printf("found %d ROM(s)\n", p.Total)
			}
		case pipeline.PhaseHashing:
			if p.Total == 0 {
				return
			}
			if p.Done == p.Total {
				fmt.Fprintf(os.Stderr, "\rhashing %d/%d\n", p.Done, p.Total)
			} else if p.Done%50 == 0 {
				fmt.Fprintf(os.Stderr, "\rhashing %d/%d", p.Done, p.Total)
			}
		case pipeline.PhaseFetching:
			if p.Done == 0 {
				fmt.Printf("fetching DB: %s\n", p.Msg)
			} else {
				fmt.Printf("%d games in DB\n", p.Total)
			}
		}
	}
}
