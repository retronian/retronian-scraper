package normalize

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FileResult struct {
	Plan      *FilePlan
	Changed   []FileAction
	NoOp      []FileAction
	Conflicts []FileAction
	Unknown   []FileAction
	Skipped   []FileAction
	Errors    []FileActionError
}

type FileActionError struct {
	Action FileAction
	Err    error
}

func (e FileActionError) Error() string {
	return fmt.Sprintf("%s %s → %s: %v", e.Action.Operation, e.Action.Source, e.Action.Target, e.Err)
}

func ApplyFilePlan(plan *FilePlan, dryRun bool, onAction func(FileAction)) (*FileResult, error) {
	if plan == nil {
		return nil, fmt.Errorf("apply file plan: nil plan")
	}
	res := &FileResult{Plan: plan}

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
				res.Changed = append(res.Changed, a)
			} else if err := applyFileAction(a); err != nil {
				res.Errors = append(res.Errors, FileActionError{Action: a, Err: err})
			} else {
				res.Changed = append(res.Changed, a)
			}
		}
		if onAction != nil {
			onAction(a)
		}
	}
	return res, nil
}

func applyFileAction(a FileAction) error {
	switch a.Operation {
	case FileOpRename:
		if strings.EqualFold(filepath.Ext(a.Source), ".zip") && a.InnerTarget != "" {
			return rewriteZipFile(a.Source, a.Target, a.InnerTarget)
		}
		return renameFile(a.Source, a.Target)
	case FileOpZip:
		return zipFile(a.Source, a.Target, a.InnerTarget)
	case FileOpUnzip:
		return unzipFile(a.Source, a.Target)
	case FileOpNoop:
		return nil
	default:
		return fmt.Errorf("unknown file operation %q", a.Operation)
	}
}

func rewriteZipFile(source, target, innerName string) error {
	zr, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer zr.Close()

	var zf *zip.File
	for _, f := range zr.File {
		if f.FileInfo().IsDir() || strings.HasPrefix(filepath.Base(f.Name), ".") {
			continue
		}
		zf = f
		break
	}
	if zf == nil {
		return fmt.Errorf("zip contains no ROM file")
	}

	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	tmp := target + fmt.Sprintf(".tmp-rewrite-%d", os.Getpid())
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	removeTmp := true
	defer func() {
		out.Close()
		if removeTmp {
			_ = os.Remove(tmp)
		}
	}()

	zw := zip.NewWriter(out)
	w, err := zw.Create(filepath.Base(innerName))
	if err != nil {
		_ = zw.Close()
		return err
	}
	if _, err := io.Copy(w, rc); err != nil {
		_ = zw.Close()
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}

	if sameCleanPath(source, target) {
		if err := os.Rename(tmp, target); err != nil {
			return err
		}
		removeTmp = false
		return nil
	}
	if err := os.Rename(tmp, target); err != nil {
		return err
	}
	if err := os.Remove(source); err != nil {
		return err
	}
	removeTmp = false
	return nil
}

func renameFile(source, target string) error {
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

func zipFile(source, target, innerName string) error {
	if innerName == "" {
		innerName = filepath.Base(source)
	}

	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(target)
	if err != nil {
		return err
	}
	removeTarget := true
	defer func() {
		out.Close()
		if removeTarget {
			_ = os.Remove(target)
		}
	}()

	zw := zip.NewWriter(out)
	w, err := zw.Create(filepath.Base(innerName))
	if err != nil {
		_ = zw.Close()
		return err
	}
	if _, err := io.Copy(w, in); err != nil {
		_ = zw.Close()
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if err := os.Remove(source); err != nil {
		return err
	}
	removeTarget = false
	return nil
}

func unzipFile(source, target string) error {
	zr, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer zr.Close()

	var zf *zip.File
	for _, f := range zr.File {
		if f.FileInfo().IsDir() || strings.HasPrefix(filepath.Base(f.Name), ".") {
			continue
		}
		zf = f
		break
	}
	if zf == nil {
		return fmt.Errorf("zip contains no ROM file")
	}

	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(target)
	if err != nil {
		return err
	}
	removeTarget := true
	defer func() {
		out.Close()
		if removeTarget {
			_ = os.Remove(target)
		}
	}()

	if _, err := io.Copy(out, rc); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if err := os.Remove(source); err != nil {
		return err
	}
	removeTarget = false
	return nil
}

func firstZipFileName(path string) (string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.FileInfo().IsDir() || strings.HasPrefix(filepath.Base(f.Name), ".") {
			continue
		}
		return filepath.Base(f.Name), nil
	}
	return "", fmt.Errorf("zip contains no ROM file")
}
