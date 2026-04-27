package normalize

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ActionStatus string

const (
	StatusRename   ActionStatus = "rename"
	StatusNoop     ActionStatus = "noop"
	StatusConflict ActionStatus = "conflict"
	StatusUnknown  ActionStatus = "unknown"
	StatusSkipped  ActionStatus = "skipped"
)

// Action describes one decision for a single subfolder of the ROM root.
type Action struct {
	Source     string // absolute path of the original folder
	Target     string // absolute path of the desired folder (== Source for noop)
	InternalID string // resolved internal platform ID; "" for unknown/skipped
	Profile    FrontendID
	Status     ActionStatus
	Detection  Detection
	Reason     string // explanation for conflict/unknown/skipped
	Fallback   bool   // true when the frontend doesn't officially support InternalID
}

// Plan is the full set of actions for one ROM root directory.
type Plan struct {
	ROMParentDir string
	Profile      Profile
	Language     LanguageID
	Actions      []Action // sorted by basename(Source)
}

// BuildPlan inspects the immediate children of romParentDir and produces
// an Action for each non-hidden entry. It performs no rename: the result
// is purely descriptive and safe to display.
func BuildPlan(romParentDir string, profile Profile) (*Plan, error) {
	return BuildPlanForLanguage(romParentDir, profile, LanguageEnglish)
}

// BuildPlanForLanguage is BuildPlan with language-aware target folder names.
func BuildPlanForLanguage(romParentDir string, profile Profile, lang LanguageID) (*Plan, error) {
	info, err := os.Stat(romParentDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s: not a directory", romParentDir)
	}

	entries, err := os.ReadDir(romParentDir)
	if err != nil {
		return nil, err
	}

	var actions []Action
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue // hidden folders are silently ignored
		}

		fullPath := filepath.Join(romParentDir, name)

		// Use Lstat so we detect symlinks before following them.
		fi, err := os.Lstat(fullPath)
		if err != nil {
			actions = append(actions, Action{
				Source:  fullPath,
				Profile: profile.ID,
				Status:  StatusSkipped,
				Reason:  fmt.Sprintf("stat error: %v", err),
			})
			continue
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			actions = append(actions, Action{
				Source:  fullPath,
				Profile: profile.ID,
				Status:  StatusSkipped,
				Reason:  "symbolic link",
			})
			continue
		}
		if !fi.IsDir() {
			continue // ignore loose files at the root
		}

		actions = append(actions, classify(romParentDir, fullPath, name, profile, lang))
	}

	sort.Slice(actions, func(i, j int) bool {
		return filepath.Base(actions[i].Source) < filepath.Base(actions[j].Source)
	})

	resolveTargetCollisions(actions)

	return &Plan{
		ROMParentDir: romParentDir,
		Profile:      profile,
		Language:     lang,
		Actions:      actions,
	}, nil
}

// classify resolves the Action for one already-validated subdirectory.
func classify(romParentDir, source, name string, profile Profile, lang LanguageID) Action {
	det, err := DetectPlatform(source, name)
	if err != nil {
		return Action{
			Source:  source,
			Profile: profile.ID,
			Status:  StatusSkipped,
			Reason:  fmt.Sprintf("detect error: %v", err),
		}
	}

	if det.Method == DetectFailed {
		return Action{
			Source:    source,
			Profile:   profile.ID,
			Status:    StatusUnknown,
			Detection: det,
			Reason:    det.Note,
		}
	}

	target, supported := profile.TargetFolderForLanguage(det.InternalID, lang)
	targetPath := filepath.Join(romParentDir, target)

	a := Action{
		Source:     source,
		Target:     targetPath,
		InternalID: det.InternalID,
		Profile:    profile.ID,
		Detection:  det,
		Fallback:   !supported,
	}

	switch {
	case filepath.Base(source) == target:
		a.Status = StatusNoop
	case targetExists(targetPath, source):
		a.Status = StatusConflict
		a.Reason = fmt.Sprintf("target already exists at %s", targetPath)
	default:
		a.Status = StatusRename
	}
	return a
}

// targetExists reports whether targetPath refers to a directory that
// already exists and is distinct from source. Same-path (case-only
// rename on case-insensitive FS) is not considered a conflict; apply.go
// handles those via a 2-step rename.
func targetExists(targetPath, source string) bool {
	fi, err := os.Lstat(targetPath)
	if err != nil {
		return false
	}
	if !fi.IsDir() {
		return false
	}
	srcInfo, err := os.Lstat(source)
	if err != nil {
		return false
	}
	return !os.SameFile(fi, srcInfo)
}

// resolveTargetCollisions marks all rename Actions sharing the same
// Target with StatusConflict so we never silently lose a folder by
// renaming two sources to the same destination.
func resolveTargetCollisions(actions []Action) {
	if len(actions) < 2 {
		return
	}
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
		for _, i := range idxs {
			actions[i].Status = StatusConflict
			actions[i].Reason = fmt.Sprintf("multiple sources map to same target %s", target)
		}
	}
}
