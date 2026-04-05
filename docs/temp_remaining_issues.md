# Remaining Issues

## HIGH

### 1. Non-atomic update destroys packages on failure

`update` コマンドは `os.RemoveAll(pkg.GetHome())` → `Install()` の順で実行する。
Install が途中で失敗するとパッケージが消えた状態で放置される。

**対処:**
- temp dir に新しいバージョンを install
- 成功したら旧ディレクトリと atomic に swap (`os.Rename`)
- 失敗したら temp dir を捨てて旧バージョンを維持

**対象ファイル:**
- `cmd/update.go` — TaskFunc 内の `os.RemoveAll` + `Install` ロジック
- `internal/runner/runner.go` — 必要なら atomic swap ヘルパー追加

### 2. No integration tests for core operations

Install/Uninstall/Clone のコア動作にテストがゼロ。
純粋関数のテスト (37%) はあるが「実際に git clone して symlink を貼る」検証がない。

**対処:**
- `internal/pkg/` に統合テスト用の TestMain + fixture を追加
- Local パッケージの Install → Installed → Uninstall のライフサイクルテスト（ネットワーク不要）
- Command.GetLink + Command.Install の symlink テスト（temp dir で完結）
- HTTP パッケージは httptest.Server でモック

**対象ファイル:**
- `internal/pkg/integration_test.go` — 新規
- `internal/pkg/command_integration_test.go` — 新規

## MEDIUM

### 3. metaCmd is still a God Object

7フィールド、150行の init()。全コマンドが embed。テスト困難。

**対処:**
- `init()` を小さな関数に分割: loadConfigs(), initPackages(), initState(), initEnv()
- metaCmd のフィールドを減らす: configs は init 後に不要なので保持しない
- コマンドの embed を interface 経由に変更

**対象ファイル:**
- `cmd/meta.go`

### 4. Hardcoded paths (~/.afx, ~/.config/afx, ~/bin)

XDG Base Directory 非対応。環境変数での変更不可。

**対処:**
- `internal/pkg/paths.go` に集約:
  - `DataDir()` — `$AFX_DATA_DIR` or `$XDG_DATA_HOME/afx` or `~/.afx`
  - `ConfigDir()` — `$AFX_CONFIG_DIR` or `$XDG_CONFIG_HOME/afx` or `~/.config/afx`
  - `BinDir()` — `$AFX_COMMAND_PATH` or `~/bin`
- 全ハードコードを上記関数に置換

**対象ファイル:**
- `internal/pkg/paths.go` — 新規
- `cmd/meta.go` — パス参照を置換
- `internal/pkg/github.go`, `gist.go`, `http.go` — GetHome() 内のパス

## LOW

### 5. internal/pkg naming is ambiguous

「pkg」はドメインを表していない。

**対処:**
- `internal/pkg/` → `internal/manager/` or `internal/installer/` にリネーム
- 全 import パスを一括置換

### 6. Custom error library is redundant

`internal/errors` は `hashicorp/go-multierror` + `pkg/errors` のラッパー。
Go 1.13+ の `errors.Join` / `fmt.Errorf("%w")` と機能が被る。

**対処:**
- `errors.Errors` → `errors.Join` に段階的に置換
- `errors.Wrapf` → `fmt.Errorf("%s: %w", msg, err)` に置換
- `internal/errors` パッケージ廃止、依存削減

### 7. mholt/archiver v3 is unmaintained

**対処:**
- `mholt/archiver/v4` への移行 or `archive/tar` + `compress/gzip` 標準ライブラリ直接利用

### 8. No file lock on state.json

複数プロセスから同時アクセスで壊れる可能性。

**対処:**
- `internal/state/` に flock ベースのファイルロック追加
- `state.Open()` でロック取得、`state.Close()` で解放

## Progress

| # | Issue | Status |
|---|-------|--------|
| 1 | Non-atomic update | **done** |
| 2 | Integration tests | **done** |
| 3 | metaCmd decomposition | **done** |
| 4 | Hardcoded paths | **done** |
| 5 | pkg naming | deferred (mechanical rename, separate PR) |
| 6 | Error library cleanup | deferred (large scope, separate PR) |
| 7 | archiver migration | deferred (external API research needed) |
| 8 | State file lock | deferred (OS portability, separate PR) |
