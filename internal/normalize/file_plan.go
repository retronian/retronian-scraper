package normalize

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/retronian/retronian-scraper/internal/export"
	"github.com/retronian/retronian-scraper/internal/match"
)

type FileFormat string

const (
	FileFormatRaw FileFormat = "raw"
	FileFormatZip FileFormat = "zip"
)

type FileOperation string

const (
	FileOpNoop   FileOperation = "noop"
	FileOpRename FileOperation = "rename"
	FileOpZip    FileOperation = "zip"
	FileOpUnzip  FileOperation = "unzip"
)

type FileOptions struct {
	ROMDir   string
	Platform string
	Profile  Profile
	Format   FileFormat
}

type FilePlan struct {
	ROMDir   string
	Platform string
	Profile  Profile
	Format   FileFormat
	Actions  []FileAction
}

type FileAction struct {
	Source      string
	Target      string
	InnerSource string
	InnerTarget string
	Profile     FrontendID
	Operation   FileOperation
	Status      ActionStatus
	Result      match.Result
	Reason      string
}

func ParseFileFormat(s string) (FileFormat, error) {
	switch FileFormat(strings.ToLower(strings.TrimSpace(s))) {
	case "", FileFormatRaw:
		return FileFormatRaw, nil
	case FileFormatZip:
		return FileFormatZip, nil
	default:
		return "", fmt.Errorf("unknown file format %q (known: raw, zip)", s)
	}
}

func BuildFilePlan(opts FileOptions, results []match.Result) (*FilePlan, error) {
	if opts.Format == "" {
		opts.Format = FileFormatRaw
	}
	if opts.Format != FileFormatRaw && opts.Format != FileFormatZip {
		return nil, fmt.Errorf("unknown file format %q", opts.Format)
	}

	actions := make([]FileAction, 0, len(results))
	for _, r := range results {
		actions = append(actions, classifyFile(opts, r))
	}
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Source < actions[j].Source
	})
	resolveFileTargetCollisions(actions)

	return &FilePlan{
		ROMDir:   opts.ROMDir,
		Platform: opts.Platform,
		Profile:  opts.Profile,
		Format:   opts.Format,
		Actions:  actions,
	}, nil
}

func classifyFile(opts FileOptions, r match.Result) FileAction {
	a := FileAction{
		Source:  r.Path,
		Profile: opts.Profile.ID,
		Result:  r,
	}
	if r.Game == nil {
		a.Status = StatusUnknown
		a.Reason = "unmatched ROM"
		return a
	}

	rawName, err := targetRawName(opts.Profile, r)
	if err != nil {
		a.Status = StatusSkipped
		a.Reason = err.Error()
		return a
	}
	dir := filepath.Dir(r.Path)
	srcIsZip := strings.EqualFold(filepath.Ext(r.Path), ".zip")

	switch opts.Format {
	case FileFormatRaw:
		a.InnerSource = zipInnerName(r.Path)
		a.Target = filepath.Join(dir, rawName)
		if srcIsZip {
			a.Operation = FileOpUnzip
		} else if sameTargetFile(r.Path, a.Target) || alreadyDisambiguatedFileName(opts.Profile, r.Path, rawName) {
			a.Target = r.Path
			a.Operation = FileOpNoop
			a.Status = StatusNoop
			return a
		} else {
			a.Operation = FileOpRename
		}
	case FileFormatZip:
		a.InnerTarget = rawName
		zipName := strings.TrimSuffix(rawName, filepath.Ext(rawName)) + ".zip"
		a.Target = filepath.Join(dir, zipName)
		if srcIsZip {
			a.InnerSource = zipInnerName(r.Path)
			if sameTargetFile(r.Path, a.Target) && a.InnerSource == a.InnerTarget {
				a.Operation = FileOpNoop
				a.Status = StatusNoop
				return a
			}
			a.Operation = FileOpRename
		} else {
			a.Operation = FileOpZip
		}
	}

	if targetFileExists(a.Target, r.Path) {
		a.Status = StatusConflict
		a.Reason = fmt.Sprintf("target already exists at %s", a.Target)
		return a
	}
	a.Status = StatusRename
	return a
}

func targetRawName(profile Profile, r match.Result) (string, error) {
	ext := filepath.Ext(r.Path)
	if strings.EqualFold(ext, ".zip") {
		inner := zipInnerName(r.Path)
		if inner == "" {
			return "", fmt.Errorf("zip contains no ROM file")
		}
		ext = filepath.Ext(inner)
	}

	if profile.ID == FrontendMinUI || profile.ID == FrontendUnuOS {
		title := export.PickTitle(r.Game.Titles)
		if title == "" {
			return "", fmt.Errorf("matched game has no title")
		}
		return sanitizeFileBase(title) + ext, nil
	}

	if r.ROM != nil && r.ROM.Name != "" {
		return sanitizeFileName(filepath.Base(r.ROM.Name)), nil
	}

	title := export.PickTitle(r.Game.Titles)
	if title == "" {
		return "", fmt.Errorf("matched game has no ROM name or title")
	}
	return sanitizeFileBase(title) + ext, nil
}

func zipInnerName(path string) string {
	name, _ := firstZipFileName(path)
	return name
}

func sanitizeFileBase(s string) string {
	return strings.TrimSpace(sanitizeFileName(s))
}

func sanitizeFileName(s string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	out := strings.TrimSpace(replacer.Replace(s))
	if out == "" || out == "." || out == ".." {
		return "unknown"
	}
	return out
}

func targetFileExists(target, source string) bool {
	fi, err := os.Lstat(target)
	if err != nil {
		return false
	}
	srcInfo, err := os.Lstat(source)
	if err != nil {
		return true
	}
	return !os.SameFile(fi, srcInfo)
}

func resolveFileTargetCollisions(actions []FileAction) {
	groups := map[string][]int{}
	for i, a := range actions {
		if a.Status != StatusRename {
			continue
		}
		groups[a.Target] = append(groups[a.Target], i)
	}
	for target, idxs := range groups {
		if len(idxs) < 2 {
			continue
		}
		if disambiguateFileTargetCollision(actions, idxs) {
			continue
		}
		for _, i := range idxs {
			actions[i].Status = StatusConflict
			actions[i].Reason = fmt.Sprintf("multiple sources map to same target %s", target)
		}
	}
}

func disambiguateFileTargetCollision(actions []FileAction, idxs []int) bool {
	seen := map[string]bool{}
	targets := make(map[int]string, len(idxs))
	innerTargets := make(map[int]string, len(idxs))
	for _, i := range idxs {
		target, innerTarget, ok := disambiguatedFileTarget(actions[i], seen)
		if !ok || targetFileExists(target, actions[i].Source) {
			return false
		}
		seen[target] = true
		targets[i] = target
		innerTargets[i] = innerTarget
	}
	for i, target := range targets {
		actions[i].Target = target
		if innerTargets[i] != "" {
			actions[i].InnerTarget = innerTargets[i]
		}
	}
	return true
}

func disambiguatedFileTarget(a FileAction, seen map[string]bool) (string, string, bool) {
	base := strings.TrimSuffix(filepath.Base(a.Target), filepath.Ext(a.Target))
	for _, suffix := range fileCollisionSuffixes(a) {
		if suffix == "" || suffix == base {
			continue
		}
		ext := filepath.Ext(a.Target)
		target := filepath.Join(filepath.Dir(a.Target), fmt.Sprintf("%s (%s)%s", base, suffix, ext))
		if seen[target] {
			continue
		}
		innerTarget := ""
		if a.InnerTarget != "" {
			innerExt := filepath.Ext(a.InnerTarget)
			innerTarget = fmt.Sprintf("%s (%s)%s", strings.TrimSuffix(a.InnerTarget, innerExt), suffix, innerExt)
		}
		return target, innerTarget, true
	}
	return "", "", false
}

func fileCollisionSuffixes(a FileAction) []string {
	var candidates []string
	if a.Result.ROM != nil && a.Result.ROM.Name != "" {
		candidates = append(candidates, strings.TrimSuffix(filepath.Base(a.Result.ROM.Name), filepath.Ext(a.Result.ROM.Name)))
	}
	sourceBase := strings.TrimSuffix(filepath.Base(a.Source), filepath.Ext(a.Source))
	candidates = append(candidates, sourceBase)
	if ext := strings.TrimPrefix(filepath.Ext(a.Source), "."); ext != "" {
		candidates = append(candidates, sourceBase+" "+ext)
	}

	seen := map[string]bool{}
	out := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		candidate = sanitizeFileBase(candidate)
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		out = append(out, candidate)
	}
	return out
}

func alreadyDisambiguatedFileName(profile Profile, path, rawName string) bool {
	if profile.ID != FrontendMinUI && profile.ID != FrontendUnuOS {
		return false
	}
	if !strings.EqualFold(filepath.Ext(path), filepath.Ext(rawName)) {
		return false
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	targetBase := strings.TrimSuffix(rawName, filepath.Ext(rawName))
	return strings.HasPrefix(base, targetBase+" (") && strings.HasSuffix(base, ")")
}

func sameCleanPath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

func sameTargetFile(a, b string) bool {
	if sameCleanPath(a, b) {
		return true
	}
	ai, err := os.Lstat(a)
	if err != nil {
		return false
	}
	bi, err := os.Lstat(b)
	if err != nil {
		return false
	}
	return os.SameFile(ai, bi)
}
