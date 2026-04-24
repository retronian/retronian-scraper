package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/retronian/retronian-scraper/internal/normalize"
)

func Normalize(args []string) int {
	return runNormalize(args, os.Stdout, os.Stderr)
}

func runNormalize(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("normalize", flag.ContinueOnError)
	fs.SetOutput(stderr)
	frontend := fs.String("frontend", "", "frontend id ("+strings.Join(normalize.KnownFrontends(), ", ")+")")
	apply := fs.Bool("apply", false, "perform rename (default: dry-run)")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: retronian-scraper normalize <rom-parent-dir> --frontend <id> [--apply]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() < 1 || *frontend == "" {
		fs.Usage()
		return 2
	}

	profile, err := normalize.LookupProfile(*frontend)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	plan, err := normalize.BuildPlan(fs.Arg(0), profile)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	mode := "dry-run"
	if *apply {
		mode = "applying"
	}
	fmt.Fprintf(stdout, "normalize: %s %s (frontend=%s)\n\n", mode, plan.ROMParentDir, profile.ID)

	res, err := normalize.Apply(plan, !*apply, nil)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	for _, a := range plan.Actions {
		fmt.Fprintln(stdout, formatAction(a, *apply))
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, formatSummary(res, *apply))
	if !*apply && hasRenames(res) {
		fmt.Fprintln(stdout, "re-run with --apply to perform rename")
	}

	if len(res.Conflicts) > 0 || len(res.Errors) > 0 {
		return 1
	}
	return 0
}

func formatAction(a normalize.Action, applied bool) string {
	src := filepath.Base(a.Source) + "/"
	switch a.Status {
	case normalize.StatusRename:
		tgt := filepath.Base(a.Target) + "/"
		tag := "rename"
		if applied {
			tag = "renamed"
		}
		if a.Fallback {
			tag += " (fallback)"
		}
		return fmt.Sprintf("  %-30s → %-25s %-10s [%s]", src, tgt, methodTag(a.Detection), tag)
	case normalize.StatusNoop:
		return fmt.Sprintf("  %-30s   %-25s %-10s [noop]", src, "", methodTag(a.Detection))
	case normalize.StatusConflict:
		return fmt.Sprintf("  %-30s   %-25s %-10s [conflict: %s]", src, "", methodTag(a.Detection), a.Reason)
	case normalize.StatusUnknown:
		return fmt.Sprintf("  %-30s   %-25s %-10s [unknown: %s]", src, "", "—", a.Reason)
	case normalize.StatusSkipped:
		return fmt.Sprintf("  %-30s   %-25s %-10s [skipped: %s]", src, "", "—", a.Reason)
	}
	return fmt.Sprintf("  %s [%s]", src, a.Status)
}

func methodTag(d normalize.Detection) string {
	switch d.Method {
	case normalize.DetectByAlias:
		return "alias"
	case normalize.DetectByContents:
		return "contents"
	}
	return "—"
}

func formatSummary(res *normalize.Result, applied bool) string {
	verb := "rename"
	if applied {
		verb = "renamed"
	}
	parts := []string{
		fmt.Sprintf("%d %s", len(res.Renamed), verb),
		fmt.Sprintf("%d noop", len(res.NoOp)),
		fmt.Sprintf("%d conflict", len(res.Conflicts)),
		fmt.Sprintf("%d unknown", len(res.Unknown)),
		fmt.Sprintf("%d skipped", len(res.Skipped)),
	}
	if len(res.Errors) > 0 {
		parts = append(parts, fmt.Sprintf("%d error", len(res.Errors)))
	}
	return "summary: " + strings.Join(parts, ", ")
}

func hasRenames(res *normalize.Result) bool {
	return len(res.Renamed) > 0
}
