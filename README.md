# Retronian Scraper

多言語対応 ROM メタデータ GUI ツール。[native-game-db](https://github.com/retronian/native-game-db) を消費し、ローカルの ROM コレクションに日本語ネイティブスクリプトのタイトルおよび boxart を適用、ES-DE / EmulationStation gamelist.xml や MinUI / OneOS 形式で書き出す。

MusicBrainz Picard が音楽ファイルに対して行うことを、レトロゲーム ROM に対してやる。

## Status

🚧 Phase 2.6 (GUI 最小版 + ROM フォルダ / ファイル名正規化)

## Build

```bash
go build -o retronian-scraper ./cmd/retronian-scraper
```

## Run

GUI を起動:

```bash
./retronian-scraper
```

CLI モード:

```bash
./retronian-scraper scan --platform gb --out gamelist.xml /path/to/roms
```

サポート対象 platform: `fc`, `sfc`, `gb`, `gbc`, `gba`, `md`, `pce`, `n64`, `nds`, `ps1`

## ROM フォルダ名の正規化

ROM 親フォルダ配下の platform サブフォルダ名 (`gb` / `Game Boy` / `GameBoy` / `GB` などバラバラなもの) を、選択した frontend の公式規約に従って一括で rename できる。

```bash
# dry-run (デフォルト): どう rename されるかだけ表示
./retronian-scraper normalize --frontend es-de /path/to/Roms

# --apply で実際に rename
./retronian-scraper normalize --frontend es-de --apply /path/to/Roms
```

サポート frontend: `es-de`, `onion` (OnionOS), `minui`, `unuui`, `batocera`, `recalbox`

例 (内部 platform ID → 各 frontend のフォルダ名):

| 内部 ID | es-de | onion | minui / unuui | batocera / recalbox |
|---|---|---|---|---|
| `gb` | `gb` | `GB` | `Game Boy (GB)` | `gb` |
| `sfc` | `snes` | `SFC` | `Super Nintendo Entertainment System (SFC)` | `snes` |
| `md` | `megadrive` | `MD` | `Sega Genesis (MD)` | `megadrive` |
| `ps1` | `psx` | `PS` | `Sony PlayStation (PS)` | `psx` |

frontend が公式に未対応の platform (例: minui の `n64`) は内部 ID をそのままフォルダ名に使う (`fallback` と表示)。

## ROM ファイル名の正規化

ROM ファイルを native-game-db で照合し、frontend に合わせてファイル名を rename できる。MinUI / UnuUI は日本語タイトル、それ以外は DB の `ROM.name` (No-Intro 由来名) を優先する。

```bash
# MinUI 向け: 日本語タイトルの raw ROM ファイル名にする
./retronian-scraper normalize --files --frontend minui --platform gb --format raw /path/to/Roms/Game\ Boy\ \(GB\)

# ES-DE 向け: No-Intro 名の zip にする
./retronian-scraper normalize --files --frontend es-de --platform gb --format zip /path/to/Roms/gb

# --apply で実際に rename / zip / unzip
./retronian-scraper normalize --files --frontend es-de --platform gb --format zip --apply /path/to/Roms/gb
```

`--format raw` は raw ROM ファイル、`--format zip` は ROM 1 本入り zip に正規化する。入力が zip の場合も zip 内 ROM をハッシュして照合する。

注意:
- ROM 親フォルダの **直下 1 階層のみ** が rename 対象 (再帰しない)
- `--files` は指定した platform の ROM ディレクトリ配下を再帰的に scan する
- 隠しフォルダ (`.git` 等) とシンボリックリンクは skip
- rename 先のフォルダが既に存在する場合は `conflict` として skip
- macOS の case-insensitive FS でも大文字小文字違いの rename (`SNES/` → `snes/`) は安全に処理
- Go の `flag` パッケージの仕様上、`<rom-parent-dir>` は flag より **後ろ** に置く必要あり

## License

Apache License 2.0 — see [LICENSE](./LICENSE).

Copyright 2026 Masaki Komagata and the Retronian Scraper contributors.
