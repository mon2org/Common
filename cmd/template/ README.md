# Go CLI Tool Template

新しい Go 製 CLI ツールを迅速に開発するためのスケルトン（テンプレート）。

フラグ解析、共通ヘルプ出力、バージョン表示、実行環境の自動検出といった CLI の基本機能は組み込み済み。

---

## ディレクトリ構成

```text
Common/
└── cmd/
    ├── internal/                # CLI 共通内部ライブラリ
    │   ├── cliutil/            # バージョン情報・Usage ヘルプの標準化
    │   ├── envutil/            # OS / WSL2 / CI / DevContainer 判定
    │   └── executil/           # 外部コマンドの存在確認・実行
    └── template/               # ★ 新規 CLI のスケルトン
        ├── main.go             # エントリーポイント & シグナル制御
        ├── config.go           # CLI フラグ定義 & バリデーション
        └── runner.go           # コアロジックの実装

```

---

## 含まれる機能

* **標準化された CLI フラグ & ヘルプ**: `-h` / `--help` で統一フォーマットの Usage を表示。
* **バージョン自動取得**: `-v` / `--version` でバージョン、短縮コミットハッシュ、ビルド日時を表示（`runtime/debug` 経由で自動解析）。
* **安全なシグナルハンドリング**: `SIGINT` (Ctrl+C) / `SIGTERM` を受信した際に安全に終了処理を行える `context.Context` のセットアップ。
* **共通ユーティリティ**: 実行環境（OS, WSL, CI等）の判定や外部コマンド実行モジュールが即座に利用可能。

---

## 新しい CLI ツールの作成手順

### 1. テンプレートをコピー

`cmd/template` ディレクトリを任意のツール名（例: `mytool`）でコピーする。

```bash
cp -r cmd/template cmd/mytool

```

### 2. メタデータの変更 (`cmd/mytool/config.go`)

`parseFlags()` 内の `cliutil.AppMeta` 情報を新しいツールの内容に書き換える。

```go
meta := cliutil.AppMeta{
    Name:        "mytool",
    Description: "ツールの概要説明をここに記述します",
    Usage:       "mytool [オプション] [引数...]",
    Examples: []string{
        "mytool --target example",
    },
}

```

### 3. フラグ・設定の追加 (`cmd/mytool/config.go`)

`Config` 構造体に必要なフィールドを定義し、`flag.StringVar` や `flag.BoolVar` 等でフラグを追加。

```go
type Config struct {
    Verbose bool
    Target  string
    Args    []string
}

// ...
flag.BoolVar(&cfg.Verbose, "verbose", false, "詳細ログを出力")
flag.StringVar(&cfg.Target, "target", "", "処理対象を指定")

```

### 4. ロジックの実装 (`cmd/mytool/runner.go`)

`run(ctx context.Context, cfg *Config) error` 関数内にメイン処理を記述する。

```go
func run(ctx context.Context, cfg *Config) error {
    env := envutil.Detect()
    
    if cfg.Verbose {
        fmt.Printf("[DEBUG] OS: %s, WSL2: %v\n", env.OS, env.IsWSL2)
    }

    // メインロジック
    fmt.Println("mytool 処理を実行中...")
    return nil
}

```

---

## ビルドと実行

### ビルド

プロジェクトルート（`Common/`）から以下のコマンドを実行し、`bin/` ディレクトリ配下にバイナリを出力。

```bash
go -C cmd build -o ../bin/mytool ./mytool

```

### 実行と動作確認

```bash
# 通常実行
./bin/mytool

# バージョン情報の表示
./bin/mytool -v
# 出力例: mytool version dev (a1b2c3d) built at 2026-08-13 11:50:00

# ヘルプの表示
./bin/mytool -h
```