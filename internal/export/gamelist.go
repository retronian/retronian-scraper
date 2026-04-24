package export

import (
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/retronian/retronian-scraper/internal/db"
	"github.com/retronian/retronian-scraper/internal/match"
)

type gameList struct {
	XMLName xml.Name `xml:"gameList"`
	Games   []game   `xml:"game"`
}

type game struct {
	Path        string `xml:"path"`
	Name        string `xml:"name"`
	Desc        string `xml:"desc,omitempty"`
	Image       string `xml:"image,omitempty"`
	ReleaseDate string `xml:"releasedate,omitempty"`
}

// WriteESDE writes an EmulationStation / ES-DE compatible gamelist.xml.
// imageDir is the relative path (e.g. "./images/") where boxart files will be placed.
func WriteESDE(w io.Writer, results []match.Result, imageDir string) error {
	if imageDir == "" {
		imageDir = "./images/"
	}
	if !strings.HasSuffix(imageDir, "/") {
		imageDir += "/"
	}

	gl := gameList{}
	for _, r := range results {
		if r.Game == nil {
			continue
		}
		g := game{
			Path: "./" + filepath.Base(r.Path),
			Name: PickTitle(r.Game.Titles),
		}
		if d := PickDescription(r.Game.Descriptions); d != "" {
			g.Desc = d
		}
		if url := PickBoxart(r.Game.Media); url != "" {
			ext := filepath.Ext(url)
			if ext == "" {
				ext = ".png"
			}
			g.Image = imageDir + r.Game.ID + ext
		}
		if d := r.Game.FirstReleaseDate; d != "" {
			g.ReleaseDate = toESDEDate(d)
		}
		gl.Games = append(gl.Games, g)
	}

	if _, err := fmt.Fprintln(w, `<?xml version="1.0" encoding="UTF-8"?>`); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(gl); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w)
	return err
}

// PickTitle selects the best display title:
//  1. verified + script in {Jpan, Hira, Kana}
//  2. any + script in {Jpan, Hira, Kana}
//  3. verified + lang=ja
//  4. first title
func PickTitle(titles []db.Title) string {
	nativeScripts := []string{"Jpan", "Hira", "Kana"}
	for _, sp := range nativeScripts {
		for _, t := range titles {
			if t.Script == sp && t.Verified {
				return t.Text
			}
		}
	}
	for _, sp := range nativeScripts {
		for _, t := range titles {
			if t.Script == sp {
				return t.Text
			}
		}
	}
	for _, t := range titles {
		if t.Lang == "ja" && t.Verified {
			return t.Text
		}
	}
	if len(titles) > 0 {
		return titles[0].Text
	}
	return ""
}

// PickDescription prefers ja descriptions, then the first available.
func PickDescription(ds []db.Description) string {
	for _, d := range ds {
		if d.Lang == "ja" {
			return d.Text
		}
	}
	if len(ds) > 0 {
		return ds[0].Text
	}
	return ""
}

// PickBoxart prefers boxart+region=jp, then any boxart, then empty.
func PickBoxart(media []db.Media) string {
	for _, m := range media {
		if m.Kind == "boxart" && m.Region == "jp" {
			return m.URL
		}
	}
	for _, m := range media {
		if m.Kind == "boxart" {
			return m.URL
		}
	}
	return ""
}

// toESDEDate converts "1992-04-27" → "19920427T000000" (ES-DE format).
// Returns input unchanged if parsing fails.
func toESDEDate(s string) string {
	if len(s) >= 10 && s[4] == '-' && s[7] == '-' {
		return s[:4] + s[5:7] + s[8:10] + "T000000"
	}
	return s
}
