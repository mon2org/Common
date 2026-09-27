# Base Container (`container/base`)

本ディレクトリは、共通利用するベースとなる Docker コンテナイメージの定義（`Dockerfile` および初期設定）を管理しています。

## 仕様・含まれるツール

* **ベースイメージ**: `debian:bookworm-slim`（ダイジェスト固定により再現性を担保）
* **パッケージマネージャー / ツールランタイム**: [`mise`](https://github.com/jdx/mise) (`v2026.9.15`)
* **サポートシェル**: `bash` および `zsh`（双方で `mise` が自動有効化されます）
* **アーキテクチャ**: マルチアーキテクチャ対応（`linux/amd64`, `linux/arm64`）

---

## ディレクトリ構成

```text
container/base/
├── Dockerfile
├── README.md
└── config/
    └── .gitkeep      # 任意の初期設定ファイルを配置可能

```

* **`config/` ディレクトリについて**:
* このディレクトリ内に配置したファイルやフォルダは、ビルド時に自動的にコンテナ内の設定ディレクトリ（`${XDG_CONFIG_HOME}` = `/root/.config`）にコピーされます。
* `mise` のグローバル設定（`config.toml`）などをコンテナにあらかじめ同梱したい場合は、この `config/` 配下に配置してください。

---

## 環境変数 & XDG Base Directory 仕様

本イメージは XDG Base Directory 仕様に準拠し、各種設定・キャッシュの配置先を整理しています。

| 変数名 | 設定値 | 用途 |
| --- | --- | --- |
| `HOME` | `/root` | ホームディレクトリ |
| `XDG_DATA_HOME` | `${HOME}/.local/share` | データ・プラグイン・ツール類（miseのシム含む） |
| `XDG_CONFIG_HOME` | `${HOME}/.config` | 設定ファイル（`config/` の内容がここに展開されます） |
| `XDG_CACHE_HOME` | `${HOME}/.cache` | キャッシュファイル |
| `XDG_STATE_HOME` | `${HOME}/.local/state` | ステートファイル |

---

## 使い方・ビルド方法

### 1. ローカルでのイメージビルド

`container/base` ディレクトリに移動し、以下のコマンドを実行します。

```bash
docker build -t base-image:latest .

```

* **異なるアーキテクチャを指定してビルドする場合（例: Apple Silicon から Intel向け）**:
```bash
docker build --platform linux/amd64 -t base-image:amd64 .

```



### 2. コンテナの起動と動作確認

ビルドしたイメージを使って、各シェルで動作確認を行うことができます。

* **`bash` で起動する場合**:
```bash
docker run --rm -it base-image:latest bash
mise --version

```


* **`zsh` で起動する場合**:
```bash
docker run --rm -it base-image:latest zsh
mise --version

```



---

## メンテナンス・バージョン更新について

`mise` のバージョンを固定しているため、バージョンを上げたい場合は `Dockerfile` 内の以下の引数（`ARG`）を変更して再ビルドしてください。

```dockerfile
ARG MISE_VERSION=v2026.9.15

```