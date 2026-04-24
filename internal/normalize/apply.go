package normalize

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Result groups Actions by outcome after Apply has run.
type Result struct {
	Plan      *Plan
	Renamed   []Action
	NoOp      []Action
	Conflicts []Action
	Unknown   []Action
	Skipped   []Action
	Errors    []ActionError
}

// ActionError pairs a rename failure with its Action.
type ActionError struct {
	Action Action
	Err    error
}

func (e ActionError) Error() string {
	return fmt.Sprintf("rename %s → %s: %v", e.Action.Source, e.Action.Target, e.Err)
}

// Apply executes plan. When dryRun is true, no rename is performed and
// the Action statuses are reported unchanged. The onAction callback,
// if non-nil, is invoked once per action after its outcome is decided.
func Apply(plan *Plan, dryRun bool, onAction func(Action)) (*Result, error) {
	if plan == nil {
		return nil, fmt.Errorf("apply: nil plan")
	}
	res := &Result{Plan: plan}

	for _, a := range plan.Actions {
		switch a.Status {
		case StatusNoop:
			res.NoOp = append(res.NoOp, a)
		case StatusConflict:
			res.Conflicts = append(res.Conflicts, a)
		case StatusUnknown:
			res.Unknown = append(res.Unknown, a)
		case StatusSkipped:
			res.Skipped = append(res.Skipped, a)
		case StatusRename:
			if dryRun {
				res.Renamed = append(res.Renamed, a)
			} else if err := renameDir(a.Source, a.Target); err != nil {
				res.Errors = append(res.Errors, ActionError{Action: a, Err: err})
			} else {
				res.Renamed = append(res.Renamed, a)
			}
		}
		if onAction != nil {
			onAction(a)
		}
	}
	return res, nil
}

// renameDir performs source → target rename, transparently handling the
// case where source and target differ only in letter case on a
// case-insensitive filesystem (macOS HFS+/APFS default, Windows). For
// those, os.Rename would be a no-op, so we route through a temporary
// name first.
func renameDir(source, target string) error {
	if source == target {
		return nil
	}
	if filepath.Dir(source) == filepath.Dir(target) {
		srcBase := filepath.Base(source)
		tgtBase := filepath.Base(target)
		if srcBase != tgtBase && strings.EqualFold(srcBase, tgtBase) {
			tmp := source + fmt.Sprintf(".tmp-rename-%d", os.Getpid())
			if err := os.Rename(source, tmp); err != nil {
				return err
			}
			return os.Rename(tmp, target)
		}
	}
	return os.Rename(source, target)
}
