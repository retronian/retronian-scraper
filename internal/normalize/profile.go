// Package normalize provides ROM platform folder name normalization for
// multiple frontends (ES-DE, OnionOS, MinUI, UnuOS, Batocera, Recalbox).
//
// The internal platform IDs (fc, sfc, gb, ...) match those used by
// internal/scan. Each Profile maps an internal ID to the folder name that
// the corresponding frontend expects under its ROM root.
package normalize

import (
	"fmt"
	"sort"
	"strings"
)

type FrontendID string
type LanguageID string

const (
	FrontendESDE     FrontendID = "es-de"
	FrontendOnion    FrontendID = "onion"
	FrontendMinUI    FrontendID = "minui"
	FrontendUnuOS    FrontendID = "unuos"
	FrontendBatocera FrontendID = "batocera"
	FrontendRecalbox FrontendID = "recalbox"

	LanguageEnglish  LanguageID = "en"
	LanguageJapanese LanguageID = "ja"
	LanguageKorean   LanguageID = "ko"
	LanguageChinese  LanguageID = "zh"
	LanguageFrench   LanguageID = "fr"
	LanguageSpanish  LanguageID = "es"
	LanguageGerman   LanguageID = "de"
)

// Profile holds the folder naming convention of a single frontend.
// Folders maps the internal platform ID to the frontend's official folder
// name. IDs absent from Folders are unsupported by the frontend; callers
// fall back to the internal ID itself via TargetFolder.
type Profile struct {
	ID               FrontendID
	DisplayName      string
	Folders          map[string]string
	LocalizedFolders map[LanguageID]map[string]string
}

// TargetFolder returns the frontend folder name for the given internal
// platform ID. The second value reports whether the frontend officially
// supports the platform; if false, the returned name is the internal ID
// itself (used as a generic fallback).
func (p Profile) TargetFolder(internalID string) (string, bool) {
	return p.TargetFolderForLanguage(internalID, LanguageEnglish)
}

// TargetFolderForLanguage returns the frontend folder name for the given
// internal platform ID and language. Unsupported languages fall back to English.
func (p Profile) TargetFolderForLanguage(internalID string, lang LanguageID) (string, bool) {
	if lang != "" && lang != LanguageEnglish {
		if folders, ok := p.LocalizedFolders[lang]; ok {
			if name, ok := folders[internalID]; ok {
				return name, true
			}
		}
	}
	if name, ok := p.Folders[internalID]; ok {
		return name, true
	}
	return internalID, false
}

// Supports reports whether the frontend has an official folder name for
// the given internal platform ID.
func (p Profile) Supports(internalID string) bool {
	_, ok := p.Folders[internalID]
	return ok
}

// Profiles holds all frontend profiles keyed by FrontendID.
//
// TODO: unuos currently mirrors minui; differentiate once the fork's
// folder convention is verified on a real device.
// TODO: recalbox currently mirrors batocera; verify against a real
// Recalbox install before treating as authoritative.
var Profiles = map[FrontendID]Profile{
	FrontendESDE: {
		ID:          FrontendESDE,
		DisplayName: "ES-DE",
		Folders: map[string]string{
			"fc":  "famicom",
			"sfc": "snes",
			"gb":  "gb",
			"gbc": "gbc",
			"gba": "gba",
			"md":  "megadrive",
			"pce": "pcengine",
			"n64": "n64",
			"nds": "nds",
			"ps1": "psx",
		},
	},
	FrontendOnion: {
		ID:          FrontendOnion,
		DisplayName: "OnionOS",
		Folders: map[string]string{
			"fc":  "FC",
			"sfc": "SFC",
			"gb":  "GB",
			"gbc": "GBC",
			"gba": "GBA",
			"md":  "MD",
			"pce": "PCE",
			// n64 unsupported on Miyoo Mini Plus hardware.
			"nds": "NDS",
			"ps1": "PS",
		},
	},
	FrontendMinUI: {
		ID:          FrontendMinUI,
		DisplayName: "MinUI",
		Folders: map[string]string{
			"fc":  "Nintendo Entertainment System (FC)",
			"sfc": "Super Nintendo Entertainment System (SFC)",
			"gb":  "Game Boy (GB)",
			"gbc": "Game Boy Color (GBC)",
			"gba": "Game Boy Advance (GBA)",
			"md":  "Sega Genesis (MD)",
			"pce": "TurboGrafx-16 (PCE)",
			// n64, nds unsupported on MinUI target hardware.
			"ps1": "Sony PlayStation (PS)",
		},
		LocalizedFolders: map[LanguageID]map[string]string{
			LanguageJapanese: {
				"fc":  "ファミリーコンピュータ (FC)",
				"sfc": "スーパーファミコン (SFC)",
				"gb":  "ゲームボーイ (GB)",
				"gbc": "ゲームボーイカラー (GBC)",
				"gba": "ゲームボーイアドバンス (GBA)",
				"md":  "メガドライブ (MD)",
				"pce": "PCエンジン (PCE)",
				"ps1": "プレイステーション (PS)",
			},
			LanguageKorean: {
				"fc":  "패밀리 컴퓨터 (FC)",
				"sfc": "슈퍼 패미컴 (SFC)",
				"gb":  "게임보이 (GB)",
				"gbc": "게임보이 컬러 (GBC)",
				"gba": "게임보이 어드밴스 (GBA)",
				"md":  "메가 드라이브 (MD)",
				"pce": "PC 엔진 (PCE)",
				"ps1": "플레이스테이션 (PS)",
			},
			LanguageChinese: {
				"fc":  "红白机 (FC)",
				"sfc": "超级任天堂 (SFC)",
				"gb":  "Game Boy (GB)",
				"gbc": "Game Boy Color (GBC)",
				"gba": "Game Boy Advance (GBA)",
				"md":  "Mega Drive (MD)",
				"pce": "PC Engine (PCE)",
				"ps1": "PlayStation (PS)",
			},
			LanguageFrench: {
				"fc":  "Nintendo Entertainment System (FC)",
				"sfc": "Super Nintendo Entertainment System (SFC)",
				"gb":  "Game Boy (GB)",
				"gbc": "Game Boy Color (GBC)",
				"gba": "Game Boy Advance (GBA)",
				"md":  "Mega Drive (MD)",
				"pce": "PC Engine (PCE)",
				"ps1": "PlayStation (PS)",
			},
			LanguageSpanish: {
				"fc":  "Nintendo Entertainment System (FC)",
				"sfc": "Super Nintendo Entertainment System (SFC)",
				"gb":  "Game Boy (GB)",
				"gbc": "Game Boy Color (GBC)",
				"gba": "Game Boy Advance (GBA)",
				"md":  "Mega Drive (MD)",
				"pce": "PC Engine (PCE)",
				"ps1": "PlayStation (PS)",
			},
			LanguageGerman: {
				"fc":  "Nintendo Entertainment System (FC)",
				"sfc": "Super Nintendo Entertainment System (SFC)",
				"gb":  "Game Boy (GB)",
				"gbc": "Game Boy Color (GBC)",
				"gba": "Game Boy Advance (GBA)",
				"md":  "Mega Drive (MD)",
				"pce": "PC Engine (PCE)",
				"ps1": "PlayStation (PS)",
			},
		},
	},
	FrontendUnuOS: {
		ID:          FrontendUnuOS,
		DisplayName: "UnuOS",
		Folders: map[string]string{
			"fc":  "Nintendo Entertainment System (FC)",
			"sfc": "Super Nintendo Entertainment System (SFC)",
			"gb":  "Game Boy (GB)",
			"gbc": "Game Boy Color (GBC)",
			"gba": "Game Boy Advance (GBA)",
			"md":  "Sega Genesis (MD)",
			"pce": "TurboGrafx-16 (PCE)",
			"ps1": "Sony PlayStation (PS)",
		},
		LocalizedFolders: map[LanguageID]map[string]string{
			LanguageJapanese: {
				"fc":  "ファミリーコンピュータ (FC)",
				"sfc": "スーパーファミコン (SFC)",
				"gb":  "ゲームボーイ (GB)",
				"gbc": "ゲームボーイカラー (GBC)",
				"gba": "ゲームボーイアドバンス (GBA)",
				"md":  "メガドライブ (MD)",
				"pce": "PCエンジン (PCE)",
				"ps1": "プレイステーション (PS)",
			},
			LanguageKorean: {
				"fc":  "패밀리 컴퓨터 (FC)",
				"sfc": "슈퍼 패미컴 (SFC)",
				"gb":  "게임보이 (GB)",
				"gbc": "게임보이 컬러 (GBC)",
				"gba": "게임보이 어드밴스 (GBA)",
				"md":  "메가 드라이브 (MD)",
				"pce": "PC 엔진 (PCE)",
				"ps1": "플레이스테이션 (PS)",
			},
			LanguageChinese: {
				"fc":  "红白机 (FC)",
				"sfc": "超级任天堂 (SFC)",
				"gb":  "Game Boy (GB)",
				"gbc": "Game Boy Color (GBC)",
				"gba": "Game Boy Advance (GBA)",
				"md":  "Mega Drive (MD)",
				"pce": "PC Engine (PCE)",
				"ps1": "PlayStation (PS)",
			},
			LanguageFrench: {
				"fc":  "Nintendo Entertainment System (FC)",
				"sfc": "Super Nintendo Entertainment System (SFC)",
				"gb":  "Game Boy (GB)",
				"gbc": "Game Boy Color (GBC)",
				"gba": "Game Boy Advance (GBA)",
				"md":  "Mega Drive (MD)",
				"pce": "PC Engine (PCE)",
				"ps1": "PlayStation (PS)",
			},
			LanguageSpanish: {
				"fc":  "Nintendo Entertainment System (FC)",
				"sfc": "Super Nintendo Entertainment System (SFC)",
				"gb":  "Game Boy (GB)",
				"gbc": "Game Boy Color (GBC)",
				"gba": "Game Boy Advance (GBA)",
				"md":  "Mega Drive (MD)",
				"pce": "PC Engine (PCE)",
				"ps1": "PlayStation (PS)",
			},
			LanguageGerman: {
				"fc":  "Nintendo Entertainment System (FC)",
				"sfc": "Super Nintendo Entertainment System (SFC)",
				"gb":  "Game Boy (GB)",
				"gbc": "Game Boy Color (GBC)",
				"gba": "Game Boy Advance (GBA)",
				"md":  "Mega Drive (MD)",
				"pce": "PC Engine (PCE)",
				"ps1": "PlayStation (PS)",
			},
		},
	},
	FrontendBatocera: {
		ID:          FrontendBatocera,
		DisplayName: "Batocera",
		Folders: map[string]string{
			"fc":  "nes",
			"sfc": "snes",
			"gb":  "gb",
			"gbc": "gbc",
			"gba": "gba",
			"md":  "megadrive",
			"pce": "pcengine",
			"n64": "n64",
			"nds": "nds",
			"ps1": "psx",
		},
	},
	FrontendRecalbox: {
		ID:          FrontendRecalbox,
		DisplayName: "Recalbox",
		Folders: map[string]string{
			"fc":  "nes",
			"sfc": "snes",
			"gb":  "gb",
			"gbc": "gbc",
			"gba": "gba",
			"md":  "megadrive",
			"pce": "pcengine",
			"n64": "n64",
			"nds": "nds",
			"ps1": "psx",
		},
	},
}

// LookupProfile resolves a frontend identifier (case-insensitive) to a
// Profile. The match is exact against the FrontendID constants.
func LookupProfile(s string) (Profile, error) {
	key := FrontendID(strings.ToLower(strings.TrimSpace(s)))
	if p, ok := Profiles[key]; ok {
		return p, nil
	}
	return Profile{}, fmt.Errorf("unknown frontend %q (known: %s)", s, strings.Join(KnownFrontends(), ", "))
}

// ParseLanguage resolves a language identifier for folder normalization.
func ParseLanguage(s string) (LanguageID, error) {
	key := LanguageID(strings.ToLower(strings.TrimSpace(s)))
	if key == "" {
		return LanguageEnglish, nil
	}
	switch key {
	case LanguageEnglish, LanguageJapanese, LanguageKorean, LanguageChinese, LanguageFrench, LanguageSpanish, LanguageGerman:
		return key, nil
	default:
		return "", fmt.Errorf("unknown language %q (known: %s)", s, strings.Join(KnownLanguages(), ", "))
	}
}

// KnownLanguages returns the sorted list of folder-name languages.
func KnownLanguages() []string {
	out := []string{
		string(LanguageEnglish),
		string(LanguageJapanese),
		string(LanguageKorean),
		string(LanguageChinese),
		string(LanguageFrench),
		string(LanguageSpanish),
		string(LanguageGerman),
	}
	sort.Strings(out)
	return out
}

// KnownFrontends returns the sorted list of supported frontend IDs.
func KnownFrontends() []string {
	out := make([]string, 0, len(Profiles))
	for id := range Profiles {
		out = append(out, string(id))
	}
	sort.Strings(out)
	return out
}
