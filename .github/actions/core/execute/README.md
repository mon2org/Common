# Core Execute (`actions/core/execute`)

`Core Execute` は、セキュリティ検証から環境構築、タスク実行（エラー分類・リトライ制御付き）、および自動 Pull Request 作成までを一貫して処理する Composite Action です。

---

## 📁 ファイル構成

```text
actions/core/execute/
├── action.yml    # Composite Action 定義
├── run.sh        # タスク実行・エラー制御スクリプト
└── README.md     # 本ドキュメント

```

---

## 🔄 処理フロー (Steps)

1. **Security Guard 実行 (`actions/core/guard`)**
実行コマンド・ランナー・イベントタイプ等の安全性検証を実施します。
2. **mise 環境構築 (`actions/wrap/mise`)**
ツールチェーン（`mise`）のインストールおよびキャッシュ有効化を行います。
3. **コード取得 (`actions/wrap/checkout`)**
対象リポジトリのコードをチェックアウトします。
4. **メインタスク実行 (`run.sh`)**
指定されたコマンドを実行し、Exit Code に基づくリトライ判定およびエラーサマリー出力を行います。
5. **Pull Request 作成 (`actions/wrap/create-pull-request`)**
`create-pr: 'true'` の場合、変更差分をもとに自動で Pull Request を生成します。

---

## 📥 入力パラメータ (Inputs)

| パラメータ | 必須 | デフォルト値 | 説明 |
| --- | --- | --- | --- |
| `run-command` | **Yes** | - | 実行するタスクコマンド |
| `runner` | No | `'ubuntu-latest-1-vcpu'` | 実行対象のランナー名（Guard 検証用） |
| `create-pr` | No | `'false'` | タスク完了後に PR を作成するかどうか (`'true'` / `'false'`) |
| `pr-branch` | No | `'chore/sync-templates'` | 作成する PR のブランチ名 |
| `pr-title` | No | `'chore: テンプレート・共通設定の自動同期'` | 作成する PR のタイトル |

---

## ⚡ タスク実行・エラー制御ロジック (`run.sh`)

`run.sh` では、実行コマンドの判定結果（Exit Code）に応じて以下のとおり制御を分岐します。

| Exit Code | エラー種別 | 挙動 | 理由 / 目的 |
| --- | --- | --- | --- |
| **`0`** | 成功 | 即時正常終了 (`exit 0`) | 正常完了 |
| **`137`** | Out of Memory (OOM) | リトライせず即時停止 | 同一環境でのリトライを回避し、`core.yml` 側の `ubuntu-latest` フォールバックジョブに即座に引き渡すため |
| **`126` / `127**` | パーミッション拒否 / コマンド不在 | リトライせず即時停止 | 設定ミスや権限不足であり、再試行しても成功しないため (Fail-Fast) |
| **その他** | 一時的エラー / その他異常 | 最大 2 回まで 5 秒間隔でリトライ | ネットワーク不調等による一発落ち（Flaky）を防止するため |

### 📊 GitHub Job Summary (`$GITHUB_STEP_SUMMARY`) への出力

タスクが失敗した場合、ログの探索を行わなくても失敗理由が把握できるよう、GitHub Actions の Job Summary に失敗コマンド・Exit Code・推定原因をフォーマット出力します。

---

## 🚀 使用方法 (Usage)

### 基本的な使い方

```yaml
- name: Execute Task
  uses: mon2org/Common/.github/actions/core/execute@main
  with:
    run-command: 'mise run sync'

```

### PR 自動作成を伴う実行

```yaml
- name: Execute Task and Create PR
  uses: mon2org/Common/.github/actions/core/execute@main
  with:
    run-command: 'mise run update-deps'
    create-pr: 'true'
    pr-branch: 'chore/update-deps'
    pr-title: 'chore: 依存関係の自動更新'

```