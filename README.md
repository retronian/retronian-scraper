# Retronian Scraper

Retronian Scraper is a multilingual ROM metadata tool. It consumes [retronian-gamedb](https://github.com/retronian/retronian-gamedb), applies native-script titles and box art metadata to local ROM collections, and exports frontend-friendly metadata.

Think of it as MusicBrainz Picard for retro game ROMs.

## Status

Phase 2.6: minimal GUI plus ROM folder and file name normalization.

## Build

```bash
go build -o retronian-scraper ./cmd/retronian-scraper
```

## Run

Start the GUI:

```bash
./retronian-scraper
```

Run a CLI scan:

```bash
./retronian-scraper scan --platform gb --out gamelist.xml /path/to/roms
```

Supported platforms: `fc`, `sfc`, `gb`, `gbc`, `gba`, `md`, `pce`, `n64`, `nds`, `ps1`

## ROM Folder Normalization

Retronian Scraper can rename platform subfolders under a ROM parent directory to match the selected frontend's folder naming rules. This handles mixed names such as `gb`, `Game Boy`, `GameBoy`, and `GB`.

```bash
# Dry run by default: show the planned renames only.
./retronian-scraper normalize --frontend es-de /path/to/Roms

# Apply the planned renames.
./retronian-scraper normalize --frontend es-de --apply /path/to/Roms

# Use localized display names for frontends that support them.
./retronian-scraper normalize --frontend minui --lang ja --apply /path/to/Roms
```

Supported frontends: `es-de`, `onion` (OnionOS), `minui`, `unuos` (UnuOS), `batocera`, `recalbox`
Supported folder-name languages: `de`, `en`, `es`, `fr`, `ja`, `ko`, `zh`

Examples:

| Internal ID | es-de | onion | MinUI / UnuOS | batocera / recalbox |
|---|---|---|---|---|
| `gb` | `gb` | `GB` | `Game Boy (GB)` | `gb` |
| `sfc` | `snes` | `SFC` | `Super Nintendo Entertainment System (SFC)` | `snes` |
| `md` | `megadrive` | `MD` | `Sega Genesis (MD)` | `megadrive` |
| `ps1` | `psx` | `PS` | `Sony PlayStation (PS)` | `psx` |

MinUI and UnuOS also support localized folder names with `--lang`, such as `ゲームボーイ (GB)` for Japanese and `게임보이 (GB)` for Korean.

If a frontend does not officially support a platform, such as `n64` on MinUI, Retronian Scraper keeps the internal ID as the folder name and marks the action as `fallback`.

## ROM File Normalization

Retronian Scraper can match ROM files against native-game-db and rename them for the selected frontend. MinUI / UnuOS prefer Japanese native-script titles; other frontends prefer the database `ROM.name` value, which is based on No-Intro naming.

```bash
# MinUI: rename to raw ROM files with Japanese native-script titles.
./retronian-scraper normalize --files --frontend minui --platform gb --format raw /path/to/Roms/Game\ Boy\ \(GB\)

# ES-DE: rename to No-Intro style one-ROM zip files.
./retronian-scraper normalize --files --frontend es-de --platform gb --format zip /path/to/Roms/gb

# Apply actual rename / zip / unzip operations.
./retronian-scraper normalize --files --frontend es-de --platform gb --format zip --apply /path/to/Roms/gb
```

`--format raw` normalizes to raw ROM files. `--format zip` normalizes to one-ROM zip files. Zip inputs are matched by the ROM inside the archive, not by the outer zip filename.

## MinUI Cover Art

MinUI supports cover art in a `.res` directory inside each platform folder. Retronian Scraper provides a helper script that downloads the best available box art for matched ROMs and writes files in MinUI's existing cover art layout:

```text
<platform folder>/.res/<ROM filename>.png
```

Example:

```text
ファミリーコンピュータ (FC)/.res/メジャーリーグ.zip.png
```

Run it with a ROM root and a GameDB API base URL:

```bash
go run ./scripts/download_minui_boxart.go /path/to/Roms https://gamedb.retronian.com
```

The downloader skips existing images, skips unmatched ROMs, and reports missing media or HTTP failures.

Notes:

- Folder normalization only scans the direct children of the ROM parent directory.
- `--files` scans ROM files recursively under the specified platform directory.
- Hidden folders such as `.git` and symbolic links are skipped.
- If the target folder already exists, the action is skipped as a `conflict`.
- Case-only renames such as `SNES/` to `snes/` are handled safely on macOS case-insensitive filesystems.
- Go's standard `flag` package requires positional arguments such as `<rom-parent-dir>` to appear after flags.

## License

Apache License 2.0. See [LICENSE](./LICENSE).

Copyright 2026 Retronian contributors.
