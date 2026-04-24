# Retronian Scraper

多言語対応 ROM メタデータ GUI ツール。[native-game-db](https://github.com/retronian/native-game-db) を消費し、ローカルの ROM コレクションに日本語ネイティブスクリプトのタイトルおよび boxart を適用、ES-DE / EmulationStation gamelist.xml や MinUI / OneOS 形式で書き出す。

MusicBrainz Picard が音楽ファイルに対して行うことを、レトロゲーム ROM に対してやる。

## Status

🚧 Phase 0 (プロジェクト基盤 + Fyne 動作検証)

## Build

```bash
go build -o retronian-scraper ./cmd/retronian-scraper
```

## Run

```bash
./retronian-scraper
```

## License

Apache License 2.0 — see [LICENSE](./LICENSE).

Copyright 2026 Masaki Komagata and the Retronian Scraper contributors.
