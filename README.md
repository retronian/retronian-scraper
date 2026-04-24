# Retronian Scraper

多言語対応 ROM メタデータ GUI ツール。[native-game-db](https://github.com/retronian/native-game-db) を消費し、ローカルの ROM コレクションに日本語ネイティブスクリプトのタイトルおよび boxart を適用、ES-DE / EmulationStation gamelist.xml や MinUI / OneOS 形式で書き出す。

MusicBrainz Picard が音楽ファイルに対して行うことを、レトロゲーム ROM に対してやる。

## Status

🚧 Phase 2 (GUI 最小版 — pipeline 抽出済み、scan / match / export を GUI から実行可能)

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

## License

Apache License 2.0 — see [LICENSE](./LICENSE).

Copyright 2026 Masaki Komagata and the Retronian Scraper contributors.
