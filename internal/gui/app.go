package gui

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/retronian/retronian-scraper/internal/export"
	"github.com/retronian/retronian-scraper/internal/match"
	"github.com/retronian/retronian-scraper/internal/pipeline"
)

var platforms = []string{"fc", "sfc", "gb", "gbc", "gba", "md", "pce", "n64", "nds", "ps1"}

type guiState struct {
	mu      sync.Mutex
	romDir  string
	results []match.Result
	output  *pipeline.Output
}

func (s *guiState) setOutput(o *pipeline.Output) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.output = o
	if o != nil {
		s.results = o.Results
	} else {
		s.results = nil
	}
}

func (s *guiState) snapshot() *pipeline.Output {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.output
}

func (s *guiState) len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.results)
}

func (s *guiState) at(i int) (match.Result, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if i < 0 || i >= len(s.results) {
		return match.Result{}, false
	}
	return s.results[i], true
}

func Run() {
	a := app.NewWithID("com.retronian.retronian-scraper")
	w := a.NewWindow("Retronian Scraper")

	state := &guiState{}

	title := widget.NewLabelWithStyle(
		"Retronian Scraper",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	subtitle := widget.NewLabel("Multilingual ROM scraper - native-game-db consumer")

	platformSel := widget.NewSelect(platforms, nil)
	platformSel.SetSelected("gb")

	romDirLabel := widget.NewLabel("(not selected)")
	romDirLabel.Wrapping = fyne.TextTruncate

	browseBtn := widget.NewButton("Select ROM Folder...", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uri == nil {
				return
			}
			state.mu.Lock()
			state.romDir = uri.Path()
			state.mu.Unlock()
			romDirLabel.SetText(uri.Path())
		}, w)
	})

	status := widget.NewLabel("Ready")
	progress := widget.NewProgressBar()
	progress.Hide()

	table := widget.NewTable(
		func() (int, int) { return state.len() + 1, 3 },
		func() fyne.CanvasObject {
			l := widget.NewLabel("")
			l.Wrapping = fyne.TextTruncate
			return l
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			lbl := o.(*widget.Label)
			if id.Row == 0 {
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				switch id.Col {
				case 0:
					lbl.SetText("File")
				case 1:
					lbl.SetText("Match")
				case 2:
					lbl.SetText("Title")
				}
				return
			}
			lbl.TextStyle = fyne.TextStyle{}
			r, ok := state.at(id.Row - 1)
			if !ok {
				lbl.SetText("")
				return
			}
			switch id.Col {
			case 0:
				lbl.SetText(filepath.Base(r.Path))
			case 1:
				lbl.SetText(tierLabel(r.Tier))
			case 2:
				if r.Game != nil {
					lbl.SetText(export.PickTitle(r.Game.Titles))
				} else {
					lbl.SetText("—")
				}
			}
		},
	)
	table.SetColumnWidth(0, 280)
	table.SetColumnWidth(1, 90)
	table.SetColumnWidth(2, 320)

	var scanBtn, exportBtn *widget.Button

	scanBtn = widget.NewButton("Scan", func() {
		state.mu.Lock()
		dir := state.romDir
		state.mu.Unlock()
		if dir == "" {
			dialog.ShowInformation("Selection Required", "Select a ROM folder.", w)
			return
		}
		platform := platformSel.Selected
		if platform == "" {
			dialog.ShowInformation("Selection Required", "Select a platform.", w)
			return
		}
		go runScan(w, platform, dir, state, table, status, progress, scanBtn, exportBtn)
	})

	exportBtn = widget.NewButton("Export gamelist.xml", func() {
		out := state.snapshot()
		if out == nil || len(out.Results) == 0 {
			dialog.ShowInformation("No Scan Results", "Run a scan first.", w)
			return
		}
		save := dialog.NewFileSave(func(uri fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uri == nil {
				return
			}
			defer uri.Close()
			if werr := export.WriteESDE(uri, out.Results, "./images/"); werr != nil {
				dialog.ShowError(werr, w)
				return
			}
			matched := out.TierCount[match.TierSHA1] + out.TierCount[match.TierSlug] + out.TierCount[match.TierHashFallback] + out.TierCount[match.TierNameFallback]
			status.SetText(fmt.Sprintf("Export complete: %s (%d entries)", uri.URI().Path(), matched))
		}, w)
		save.SetFileName("gamelist.xml")
		save.Show()
	})
	exportBtn.Disable()

	controls := container.NewVBox(
		container.NewBorder(nil, nil, widget.NewLabel("Platform:"), nil, platformSel),
		container.NewBorder(nil, nil, widget.NewLabel("ROM Folder:"), browseBtn, romDirLabel),
		container.NewHBox(scanBtn, exportBtn),
		progress,
	)

	top := container.NewVBox(
		title, subtitle,
		widget.NewSeparator(),
		controls,
		widget.NewSeparator(),
	)
	bottom := container.NewVBox(widget.NewSeparator(), status)
	content := container.NewBorder(top, bottom, nil, nil, table)

	w.SetContent(content)
	w.Resize(fyne.NewSize(820, 600))
	w.ShowAndRun()
}

func runScan(
	w fyne.Window,
	platform, romDir string,
	state *guiState,
	table *widget.Table,
	status *widget.Label,
	progress *widget.ProgressBar,
	scanBtn, exportBtn *widget.Button,
) {
	fyne.Do(func() {
		scanBtn.Disable()
		exportBtn.Disable()
		state.setOutput(nil)
		table.Refresh()
		progress.Show()
		progress.SetValue(0)
		status.SetText("Starting scan...")
	})

	output, err := pipeline.Run(context.Background(), pipeline.Options{
		ROMDir:   romDir,
		Platform: platform,
	}, func(p pipeline.Progress) {
		fyne.Do(func() {
			updateProgress(progress, status, p)
		})
	})

	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(err, w)
			scanBtn.Enable()
			progress.Hide()
			status.SetText("Error: " + err.Error())
		})
		return
	}

	state.setOutput(output)

	matched := output.TierCount[match.TierSHA1] + output.TierCount[match.TierSlug] + output.TierCount[match.TierHashFallback] + output.TierCount[match.TierNameFallback]
	summary := fmt.Sprintf(
		"Matched %d/%d (sha1=%d, hash=%d, name=%d, unmatched=%d)",
		matched, len(output.Results),
		output.TierCount[match.TierSHA1],
		output.TierCount[match.TierHashFallback],
		output.TierCount[match.TierNameFallback],
		output.TierCount[match.TierNone],
	)

	fyne.Do(func() {
		progress.SetValue(1.0)
		progress.Hide()
		status.SetText(summary)
		table.Refresh()
		scanBtn.Enable()
		if matched > 0 {
			exportBtn.Enable()
		}
	})
}

// updateProgress maps pipeline phases onto the bottom progress bar.
// Hashing uses 0.0-0.6, DB fetch 0.6-0.7, and matching 0.7-1.0.
func updateProgress(progress *widget.ProgressBar, status *widget.Label, p pipeline.Progress) {
	switch p.Phase {
	case pipeline.PhaseWalking:
		if p.Total > 0 {
			status.SetText(fmt.Sprintf("ROMs found: %d", p.Total))
		} else {
			status.SetText("Scanning ROM directory...")
		}
	case pipeline.PhaseHashing:
		if p.Total > 0 {
			status.SetText(fmt.Sprintf("Hashing %d/%d", p.Done, p.Total))
			progress.SetValue(float64(p.Done) / float64(p.Total) * 0.6)
		}
	case pipeline.PhaseFetching:
		if p.Done == 0 {
			status.SetText("Fetching DB: " + p.Msg)
			progress.SetValue(0.65)
		} else {
			status.SetText(fmt.Sprintf("Fetched %d DB entries", p.Total))
			progress.SetValue(0.7)
		}
	case pipeline.PhaseMatching:
		if p.Total > 0 {
			status.SetText(fmt.Sprintf("Matching %d/%d", p.Done, p.Total))
			progress.SetValue(0.7 + float64(p.Done)/float64(p.Total)*0.3)
		}
	case pipeline.PhaseDone:
		progress.SetValue(1.0)
	}
}

func tierLabel(t match.Tier) string {
	switch t {
	case match.TierSHA1:
		return "SHA1"
	case match.TierSlug:
		return "slug"
	case match.TierHashFallback:
		return "hash"
	case match.TierNameFallback:
		return "name"
	default:
		return "—"
	}
}
