package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/retronian/retronian-scraper/internal/db"
	"github.com/retronian/retronian-scraper/internal/match"
	"github.com/retronian/retronian-scraper/internal/scan"
)

func TestPickTitle_PrefersVerifiedNativeScript(t *testing.T) {
	titles := []db.Title{
		{Text: "Kirby's Dream Land", Lang: "en", Script: "Latn"},
		{Text: "星のカービィ", Lang: "ja", Script: "Jpan", Verified: true},
		{Text: "ほしのカービィ", Lang: "ja", Script: "Hira"},
	}
	got := PickTitle(titles)
	if got != "星のカービィ" {
		t.Errorf("want 星のカービィ, got %q", got)
	}
}

func TestPickTitle_FallsBackToNonVerifiedNative(t *testing.T) {
	titles := []db.Title{
		{Text: "Kirby's Dream Land", Lang: "en", Script: "Latn"},
		{Text: "星のカービィ", Lang: "ja", Script: "Jpan"},
	}
	got := PickTitle(titles)
	if got != "星のカービィ" {
		t.Errorf("want 星のカービィ, got %q", got)
	}
}

func TestPickTitle_FallsBackToFirst(t *testing.T) {
	titles := []db.Title{
		{Text: "Only Latin", Lang: "en", Script: "Latn"},
	}
	got := PickTitle(titles)
	if got != "Only Latin" {
		t.Errorf("want Only Latin, got %q", got)
	}
}

func TestPickBoxart_PrefersJP(t *testing.T) {
	media := []db.Media{
		{Kind: "boxart", Region: "us", URL: "https://example.com/us.png"},
		{Kind: "boxart", Region: "jp", URL: "https://example.com/jp.png"},
		{Kind: "titlescreen", Region: "jp", URL: "https://example.com/title.png"},
	}
	got := PickBoxart(media)
	if got != "https://example.com/jp.png" {
		t.Errorf("want jp boxart, got %q", got)
	}
}

func TestPickBoxart_FallsBackToAnyRegion(t *testing.T) {
	media := []db.Media{
		{Kind: "boxart", Region: "us", URL: "https://example.com/us.png"},
	}
	got := PickBoxart(media)
	if got != "https://example.com/us.png" {
		t.Errorf("want us boxart, got %q", got)
	}
}

func TestWriteESDE_RendersMatchedGames(t *testing.T) {
	results := []match.Result{
		{
			Path:   "/roms/gb/Kirby.gb",
			Hashes: scan.Hashes{SHA1: "deadbeef"},
			Tier:   match.TierSHA1,
			Game: &db.Game{
				ID:               "hoshi-no-kirby",
				Platform:         "gb",
				FirstReleaseDate: "1992-04-27",
				Titles: []db.Title{
					{Text: "星のカービィ", Lang: "ja", Script: "Jpan", Verified: true},
				},
				Media: []db.Media{
					{Kind: "boxart", Region: "jp", URL: "https://example.com/hoshi-no-kirby.png"},
				},
				Descriptions: []db.Description{
					{Lang: "ja", Text: "カービィの冒険"},
				},
			},
		},
		{
			Path: "/roms/gb/unknown.gb",
			Tier: match.TierNone,
		},
	}

	var buf bytes.Buffer
	if err := WriteESDE(&buf, results, "./images/"); err != nil {
		t.Fatalf("WriteESDE: %v", err)
	}
	xml := buf.String()

	wantContains := []string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<gameList>`,
		`<path>./Kirby.gb</path>`,
		`<name>星のカービィ</name>`,
		`<desc>カービィの冒険</desc>`,
		`<image>./images/hoshi-no-kirby.png</image>`,
		`<releasedate>19920427T000000</releasedate>`,
		`</gameList>`,
	}
	for _, s := range wantContains {
		if !strings.Contains(xml, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, xml)
		}
	}

	// Unmatched ROM must NOT appear
	if strings.Contains(xml, "unknown.gb") {
		t.Errorf("unmatched ROM leaked into output:\n%s", xml)
	}
}
