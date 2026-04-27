package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/retronian/retronian-scraper/internal/db"
	"github.com/retronian/retronian-scraper/internal/normalize"
	"github.com/retronian/retronian-scraper/internal/pipeline"
)

func Normalize(args []string) int {
	return runNormalize(args, os.Stdout, os.Stderr)
}

func runNormalize(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("normalize", flag.ContinueOnError)
	fs.SetOutput(stderr)
	frontend := fs.String("frontend", "", "frontend id ("+strings.Join(normalize.KnownFrontends(), ", ")+")")
	langFlag := fs.String("lang", "en", "folder name language ("+strings.Join(normalize.KnownLanguages(), ", ")+")")
	apply := fs.Bool("apply", false, "perform rename (default: dry-run)")
	files := fs.Bool("files", false, "normalize ROM filenames instead of platform folders")
	platform := fs.String("platform", "", "platform id for --files")
	format := fs.String("format", "raw", "file output format for --files (raw, zip)")
	baseURL := fs.String("api", db.DefaultBaseURL, "native-game-db API base URL for --files")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: retronian-scraper normalize <dir> --frontend <id> [--lang <lang>] [--files --platform <id> --format raw|zip] [--apply]")
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
	lang, err := normalize.ParseLanguage(*langFlag)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if *files {
		return runNormalizeFiles(fs.Arg(0), profile, *platform, *format, *baseURL, *apply, stdout, stderr)
	}

	plan, err := normalize.BuildPlanForLanguage(fs.Arg(0), profile, lang)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	mode := "dry-run"
	if *apply {
		mode = "applying"
	}
	fmt.Fprintf(stdout, "normalize: %s %s (frontend=%s, lang=%s)\n\n", mode, plan.ROMParentDir, profile.ID, lang)

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

func runNormalizeFiles(romDir string, profile normalize.Profile, platformID, formatID, baseURL string, apply bool, stdout, stderr io.Writer) int {
	if platformID == "" {
		fmt.Fprintln(stderr, "--platform is required with --files")
		return 2
	}
	format, err := normalize.ParseFileFormat(formatID)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	mode := "dry-run"
	if apply {
		mode = "applying"
	}
	fmt.Fprintf(stdout, "normalize files: %s %s (frontend=%s, platform=%s, format=%s)\n\n", mode, romDir, profile.ID, platformID, format)

	output, err := pipeline.Run(context.Background(), pipeline.Options{
		ROMDir:   romDir,
		Platform: platformID,
		BaseURL:  baseURL,
	}, nil)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	plan, err := normalize.BuildFilePlan(normalize.FileOptions{
		ROMDir:   romDir,
		Platform: platformID,
		Profile:  profile,
		Format:   format,
	}, output.Results)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	res, err := normalize.ApplyFilePlan(plan, !apply, nil)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	for _, a := range plan.Actions {
		fmt.Fprintln(stdout, formatFileAction(a, apply))
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, formatFileSummary(res, apply))
	if !apply && len(res.Changed) > 0 {
		fmt.Fprintln(stdout, "re-run with --apply to perform file changes")
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

func formatFileAction(a normalize.FileAction, applied bool) string {
	src := filepath.Base(a.Source)
	switch a.Status {
	case normalize.StatusRename:
		tgt := filepath.Base(a.Target)
		tag := string(a.Operation)
		if applied {
			switch a.Operation {
			case normalize.FileOpZip:
				tag = "zipped"
			case normalize.FileOpUnzip:
				tag = "unzipped"
			default:
				tag = "renamed"
			}
		}
		return fmt.Sprintf("  %-42s → %-42s [%s]", src, tgt, tag)
	case normalize.StatusNoop:
		return fmt.Sprintf("  %-42s   %-42s [noop]", src, "")
	case normalize.StatusConflict:
		return fmt.Sprintf("  %-42s   %-42s [conflict: %s]", src, "", a.Reason)
	case normalize.StatusUnknown:
		return fmt.Sprintf("  %-42s   %-42s [unknown: %s]", src, "", a.Reason)
	case normalize.StatusSkipped:
		return fmt.Sprintf("  %-42s   %-42s [skipped: %s]", src, "", a.Reason)
	}
	return fmt.Sprintf("  %s [%s]", src, a.Status)
}

func formatFileSummary(res *normalize.FileResult, applied bool) string {
	verb := "change"
	if applied {
		verb = "changed"
	}
	parts := []string{
		fmt.Sprintf("%d %s", len(res.Changed), verb),
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
