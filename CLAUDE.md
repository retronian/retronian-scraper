# Retronian Scraper — Claude / AI Handoff

このファイルは Claude Code (および互換 AI assistant) がこのプロジェクトで作業する際のコンテキストを与える。新規セッションを開始する AI は **必ず最初に読み込むこと**。

## プロジェクト概要

**Retronian Scraper** = 多言語対応 ROM メタデータ GUI ツール (Go + Fyne)。姉妹プロジェクト [native-game-db](https://github.com/retronian/native-game-db) の公開 JSON API を消費し、ローカル ROM コレクションに日本語ネイティブスクリプトのタイトル + boxart を適用、ES-DE / EmulationStation 用 `gamelist.xml` や MinUI / OneOS 形式で書き出す。MusicBrainz Picard の ROM 版がコンセプト。

旧名: `Babel Librarian` → `Retronian Scraper` に改名済み (2026-04-23)。

## リポジトリ

- 本 repo: `/Users/komagata/Works/retronian/retronian-scraper/`
- GitHub: https://github.com/retronian/retronian-scraper (public, Apache-2.0)
- 姉妹 repo: `/Users/komagata/Works/retronian/native-game-db/` (Ruby 製、データソース)

## 技術スタック

- 言語: **Go 1.26.2**
- GUI: **Fyne v2.7.3** (CGO 必要、OpenGL ベース)
- 配布: `go build` で single static binary (現状 33MB)
- DB アクセス: `https://gamedb.retronian.com/api/v1/{platform}.json` (JSON、ライブで叩ける)
- マッチング: 3-tier (SHA1 一次 / slug 二次 (未実装) / CRC32+MD5 fallback)
- 対象 OS: Linux / Windows / macOS
- 対象 ROM platform (10 種): `fc, sfc, gb, gbc, gba, md, pce, n64, nds, ps1`

## 主要パス

```
cmd/retronian-scraper/main.go        エントリ (CLI/GUI dispatch)
internal/scan/                       ROM walker + 並列ハッシュ計算
internal/db/                         native-game-db API クライアント
internal/match/                      3-tier matcher
internal/export/                     ES-DE gamelist.xml exporter
internal/pipeline/                   scan → hash → fetch → match の共通パイプライン
internal/cli/                        scan / normalize CLI サブコマンド
internal/gui/                        Fyne GUI
internal/normalize/                  ROM フォルダ名正規化 (6 frontend)
```

## 実装フェーズ

### Phase 0 ✅ プロジェクト基盤 (commit `f833f84`)
- Go module init、Fyne hello-world で macOS の CJK 表示を検証

### Phase 1 ✅ CLI core pipeline (commit `9a31d6b`)
- ROM walker、並列 SHA1/MD5/CRC32 hasher
- native-game-db API client
- 3-tier matcher (SHA1 / hash fallback)
- ES-DE gamelist.xml exporter (`PickTitle` は verified Jpan > Hira/Kana > ja > first)
- `scan` CLI サブコマンド

### Phase 2 ✅ GUI 最小版 + pipeline 抽出 (commit `fa53a6e`)
- `internal/pipeline/` 抽出 (CLI/GUI 共通)
- Fyne GUI: platform 選択 / ROM フォルダ選択 / scan ボタン / 進捗バー / 結果テーブル / gamelist.xml 書き出し

### Phase 2.5 ✅ ROM フォルダ名正規化 (commit `552c15a`)
- `internal/normalize/` 新規パッケージ
- 6 frontend (es-de / onion / minui / unuui / batocera / recalbox) の公式フォルダ名にリネーム
- alias 主・拡張子分布補助で内部 ID 判定
- `normalize` CLI サブコマンド (dry-run デフォルト + `--apply`)
- ユニットテスト 30 件グリーン

### 次のタスク候補

1. **実 ROM で検証** — 手元の本物 ROM で `scan` の matched N/M を確認、不具合修正
2. **Phase 3** — 手動 ID override、boxart サムネプレビュー、region 切替
3. **Phase 4** — GitHub Actions で 3 OS リリースバイナリ自動ビルド
4. **MinUI / OnionOS export** — 現状 ES-DE のみ
5. **GUI に normalize 統合** — Phase 2.5 の `normalize.BuildPlan` / `Apply` をプレビュー UI で

## ビルド・テスト・実行

```bash
# ビルド (CGO 必要、Xcode CLI tools が入ってれば OK)
go build -o retronian-scraper ./cmd/retronian-scraper

# 全パッケージのビルド・vet・テスト
go build ./... && go vet ./... && go test ./...

# GUI 起動
./retronian-scraper

# CLI scan
./retronian-scraper scan --platform gb --out gamelist.xml /path/to/roms

# CLI normalize (dry-run)
./retronian-scraper normalize --frontend es-de /path/to/Roms

# CLI normalize (実行)
./retronian-scraper normalize --frontend es-de --apply /path/to/Roms
```

**重要**: Go 標準 `flag` パッケージの仕様で **positional 引数は flag より後ろ** に置く必要がある。`./retronian-scraper scan /path --platform gb` は parse 失敗するので注意。

## 既知の未解決事項 / 注意点

- **slug tier (Tier 2)**: `internal/match/matcher.go` にコメントだけあり未実装。SHA1 でほぼマッチするので優先度低だが ROM hacker 向け collection だと効く
- **CGO 依存**: Fyne が OpenGL 使用のため CGO 必要。Linux ビルドは X11 dev libs、Windows は MinGW 要。GitHub Actions で 3 OS runner 推奨
- **macOS Gatekeeper / notarize**: 未対応。初回起動で「開発元未確認」警告が出る。README に対処手順を書くか Apple Developer Program を購入するか判断が必要
- **`.app` bundle と `Icon.png`**: プレースホルダ。後でデザインしたアイコンに差し替え必要
- **ライセンス attribution**: native-game-db データは Wikipedia 由来 (CC BY-SA 4.0)。`gamelist.xml` に attribution、GUI About に表記する TODO あり
- **deprecated `fyne` CLI**: `go install fyne.io/fyne/v2/cmd/fyne@latest` は deprecated 警告。新しい `fyne.io/tools/cmd/fyne` への移行を推奨
- **unuui / recalbox の normalize profile**: 現状暫定 (minui / batocera と同一)。実機検証後に分離する TODO コメントあり (`internal/normalize/profile.go`)
- **MinUI / OnionOS の n64 (および MinUI の nds)**: frontend 公式に未対応 → fallback で内部 ID をフォルダ名に使う

## コーディング・コミュニケーション規約

- **会話言語**: 日本語 (komagata さんの明示指示)
- **コードコメント**: 最小限。WHY が非自明な場合のみ。WHAT は自明な命名で済ませる
- **README、user-facing 文言**: 日本語 OK
- **commit message**: 英語、`type: subject` 形式 (`feat:`, `fix:`, `chore:`, `refactor:`, `docs:`)、本文で詳細
- **ブランチ**: 当面 `main` 直 commit (1 人開発、PR フローはまだ無い)
- **commit 単位**: 機能ごとに 1 commit。Phase 区切りで段階的に push 済み

## ユーザー (komagata さん) について

- レトロゲーム × 日本語ネイティブスクリプト対応に詳しい (retronian/native-game-db の主要コントリビュータ)
- `retronian` GitHub org のオーナー
- `gh` CLI に retronian / komagata 両アカウントを保持、active は retronian

## セッション開始時の確認事項

1. 本ファイル (`CLAUDE.md`) を読む
2. `git log --oneline` で最新 commit を把握
3. `git status` で進行中の作業がないか確認
4. komagata さんに「次のタスク候補」から優先順を確認

## 参考: native-game-db (姉妹プロジェクト)

- Ruby 製の静的 API generator
- `data/`: ゲームデータ (YAML)
- `schema/game.schema.json`: platform ID enum (本 repo の 10 ID と同一)
- `scripts/build_api.rb`: `PLATFORMS` ハッシュで platform ID と表示名を持つ
- 公開 API: `https://gamedb.retronian.com/api/v1/{platform}.json`
- ライセンス: dual (CODE = MIT, DATA = CC BY-SA 4.0)
