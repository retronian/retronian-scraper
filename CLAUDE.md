# Retronian Scraper - Claude / AI Handoff

This file gives Claude Code and compatible AI assistants project context. New sessions should read it before making changes.

## Project Overview

**Retronian Scraper** is a multilingual ROM metadata GUI tool built with Go and Fyne. It consumes the public JSON API from the sister project [retronian-gamedb](https://github.com/retronian/retronian-gamedb), applies native-script titles and box art metadata to local ROM collections, and exports frontend-friendly metadata. The concept is MusicBrainz Picard for ROMs.

Former name: `Babel Librarian`. Renamed to `Retronian Scraper` on 2026-04-23.

## Repositories

- This repo: `/home/komagata/Works/retronian/retronian-scraper/`
- GitHub: https://github.com/retronian/retronian-scraper (public, Apache-2.0)
- Sister repo: `/home/komagata/Works/retronian/retronian-gamedb/` (Ruby data source)

## Stack

- Language: **Go 1.26.2**
- GUI: **Fyne v2.7.3** (requires CGO, OpenGL based)
- Distribution: single binary through `go build`
- DB access: `https://gamedb.retronian.com/api/v1/{platform}.json`
- Matching: 3-tier matching (SHA1 primary / slug secondary, not implemented / CRC32+MD5 fallback)
- Target OS: Linux / Windows / macOS
- Target ROM platforms: `fc`, `sfc`, `gb`, `gbc`, `gba`, `md`, `pce`, `n64`, `nds`, `ps1`

## Main Paths

```text
cmd/retronian-scraper/main.go        CLI/GUI entry point
internal/scan/                       ROM walker and parallel hashing
internal/db/                         native-game-db API client
internal/match/                      3-tier matcher
internal/export/                     ES-DE gamelist.xml exporter
internal/pipeline/                   shared scan -> hash -> fetch -> match pipeline
internal/cli/                        scan / normalize CLI subcommands
internal/gui/                        Fyne GUI
internal/normalize/                  ROM folder and file name normalization
```

## Implementation Phases

### Phase 0 Complete: Project Foundation

- Initialized the Go module.
- Verified CJK rendering with a Fyne hello-world build.

### Phase 1 Complete: CLI Core Pipeline

- ROM walker and parallel SHA1/MD5/CRC32 hasher.
- native-game-db API client.
- 3-tier matcher (SHA1 / hash fallback currently implemented).
- ES-DE `gamelist.xml` exporter. `PickTitle` order is verified Jpan > Hira/Kana > ja > first.
- `scan` CLI subcommand.

### Phase 2 Complete: Minimal GUI and Pipeline Extraction

- Extracted `internal/pipeline/` for shared CLI/GUI use.
- Added a Fyne GUI with platform selection, ROM folder picker, scan button, progress bar, results table, and `gamelist.xml` export.

### Phase 2.5 Complete: ROM Folder Name Normalization

- Added `internal/normalize/`.
- Supports official folder-name profiles for `es-de`, `onion`, `minui`, `unuui`, `batocera`, and `recalbox`.
- Supports localized folder names through `normalize --lang <lang>` where the frontend profile provides them. Current folder-name languages are `de`, `en`, `es`, `fr`, `ja`, `ko`, and `zh`.
- Detects internal platform IDs through aliases with extension-distribution assistance.
- Added the `normalize` CLI subcommand with dry-run default and `--apply`.

### Current Status

Phase 2.6: minimal GUI plus ROM folder and file name normalization.

## Build, Test, Run

```bash
# Build
go build -o retronian-scraper ./cmd/retronian-scraper

# Build, vet, and test all packages
go build ./... && go vet ./... && go test ./...

# Start the GUI
./retronian-scraper

# CLI scan
./retronian-scraper scan --platform gb --out gamelist.xml /path/to/roms

# CLI normalize dry run
./retronian-scraper normalize --frontend es-de /path/to/Roms

# CLI normalize apply
./retronian-scraper normalize --frontend es-de --apply /path/to/Roms

# CLI normalize with localized folder names
./retronian-scraper normalize --frontend minui --lang ja --apply /path/to/Roms
```

Important: Go's standard `flag` package requires positional arguments to appear after flags. `./retronian-scraper scan /path --platform gb` will not parse as intended.

## Known Issues and Notes

- **Slug tier (Tier 2)**: referenced in `internal/match/matcher.go` but not implemented. SHA1 handles most matches, so this is lower priority unless ROM-hacker collections become important.
- **CGO dependency**: Fyne uses OpenGL and requires CGO. Linux builds need X11 dev libraries; Windows builds need MinGW. GitHub Actions should use three OS runners.
- **macOS Gatekeeper / notarization**: not handled yet. First launch may show an unidentified developer warning.
- **`.app` bundle and `Icon.png`**: placeholders that need replacement with final design assets.
- **License attribution**: native-game-db data derives from Wikipedia under CC BY-SA 4.0. `gamelist.xml` and GUI About attribution are still TODO.
- **Deprecated `fyne` CLI**: `go install fyne.io/fyne/v2/cmd/fyne@latest` shows a deprecation warning. Prefer migrating to `fyne.io/tools/cmd/fyne`.
- **`unuui` / `recalbox` normalize profiles**: currently provisional and aligned with `minui` / `batocera`. Split after hardware verification if needed.
- **MinUI / OnionOS unsupported platforms**: `n64` and MinUI `nds` fall back to the internal ID as the folder name.

## Coding and Communication Rules

- Conversation language with komagata: Japanese unless asked otherwise.
- Code comments: keep them minimal. Use comments for non-obvious why, not obvious what.
- README and user-facing text: English.
- Commit messages: English, `type: subject` format such as `feat:`, `fix:`, `chore:`, `refactor:`, or `docs:`.
- Branching: direct commits to `main` for now.
- Commit scope: one feature or coherent change per commit.

## User Context

- komagata is familiar with retro games and Japanese native-script metadata.
- komagata owns the `retronian` GitHub organization.
- The `gh` CLI may have both `retronian` and `komagata` accounts; the active account should be `retronian`.

## Session Start Checklist

1. Read this file.
2. Run `git log --oneline` to understand recent commits.
3. Run `git status` to check for in-progress work.
4. Continue from the user's latest instruction.

## retronian-gamedb Reference

- Ruby static API generator.
- `data/`: game data in YAML.
- `schema/game.schema.json`: platform ID enum, matching this repo's 10 platform IDs.
- `scripts/build_api.rb`: has the `PLATFORMS` mapping with platform IDs and display names.
- Public API: `https://gamedb.retronian.com/api/v1/{platform}.json`
- License: CODE = MIT, DATA = CC BY-SA 4.0.
