# mon2org/Common 共通 CI/CD 実行基盤 (Core System) 仕様書

## 1. 概要と基本理念

Core システムは、組織内（`mon2org`）の全リポジトリにおける CI/CD 処理の標準化・安全性の向上・運用コスト最適化を目的に開発された実行基盤です。

### 主な特徴

* **デュアル実行モードのサポート:** `mise` 経由の CLI コマンド実行（`run-command`）と、個別 Composite Action の呼び出し（`uses-action`）の両方に対応。
* **多層防御（Security Guard）:** Poisoned Pipeline Execution (PPE) 対策として、外部 Fork からの PR 拒否、コマンド/アクションの構文検証、許可されたランナー（ホワイトリスト）チェックを実施。
* **コスト最適化と自動フォールバック:** 低スペックランナー（`ubuntu-latest-1-vcpu`）で失敗（OOM 等）した場合、自動的に標準ランナー（`ubuntu-latest`）へ引き継いで再実行。
* **スマートリトライと詳細サマリー:** 一時的エラーのリトライと確定エラー（OOM/コマンド不在等）の即時停止を識別し、`$GITHUB_STEP_SUMMARY` へ失敗原因を出力。
* **重複防止付き障害通知:** メイン・フォールバックの双方が失敗した場合、既存の Open Issue を検索してコメント追記（無ければ新規 Issue 作成）を行う冪等な失敗通知。

---

## 2. ディレクトリ構造

```text
.github/
├── workflows/
│   └── core.yml                  # 共通 Reusable Workflow
└── actions/
    ├── core/
    │   ├── execute/
    │   │   ├── action.yml        # パイプライン制御 Composite Action
    │   │   └── run.sh            # タスク実行・スマートリトライ制御スクリプト
    │   ├── guard/
    │   │   ├── action.yml        # セキュリティ検証 Composite Action
    │   │   ├── checker.sh        # セキュリティ検証スクリプト
    │   │   └── README.md
    │   └── notify/
    │       ├── action.yml        # 失敗通知 Composite Action
    │       ├── notify.sh         # Issue 重複チェック・作成スクリプト
    │       └── README.md
    └── wrap/
        ├── checkout/
        │   └── action.yml        # actions/checkout ラッパー
        ├── mise/
        │   └── action.yml        # jdx/mise-action ラッパー
        └── create-pull-request/
            └── action.yml        # peter-evans/create-pull-request ラッパー

```

---

## 3. 再利用可能ワークフロー (`workflows/core.yml`)

呼び出し元リポジトリから `jobs.<job_id>.uses` で呼び出される基幹ワークフローです。

### ソースコード

```yaml
name: Core Reusable Workflow

on:
  workflow_call:
    inputs:
      run-command:
        description: '実行する Command（mise ...）'
        required: false
        type: string
        default: ''
      uses-action:
        description: '実行する Action（./.github/actions/... や mon2org/...）'
        required: false
        type: string
        default: ''
      runner:
        required: false
        type: string
        default: 'ubuntu-latest-1-vcpu'
      create-pr:
        required: false
        type: boolean
        default: false
      pr-branch:
        required: false
        type: string
        default: 'chore/sync-templates'
      pr-title:
        required: false
        type: string
        default: 'chore: テンプレート・共通設定の自動同期'
      timeout-minutes:
        required: false
        type: number
        default: 15

jobs:
  execute:
    runs-on: ${{ inputs.runner }}
    timeout-minutes: ${{ inputs.timeout-minutes }}
    steps:
      - name: Execute core process
        uses: mon2org/Common/.github/actions/core/execute@main
        with:
          run-command: ${{ inputs.run-command }}
          uses-action: ${{ inputs.uses-action }}
          runner: ${{ inputs.runner }}
          create-pr: ${{ inputs.create-pr }}
          pr-branch: ${{ inputs.pr-branch }}
          pr-title: ${{ inputs.pr-title }}

  execute-fallback:
    needs: execute
    if: failure() && inputs.runner == 'ubuntu-latest-1-vcpu'
    runs-on: ubuntu-latest
    timeout-minutes: ${{ inputs.timeout-minutes }}
    steps:
      - name: Log fallback attempt
        run: echo "⚠️ 1-vCPU ランナーでの実行に失敗したため、ubuntu-latest (2 vCPU) で再試行します。"

      - name: Execute core process
        uses: mon2org/Common/.github/actions/core/execute@main
        with:
          run-command: ${{ inputs.run-command }}
          uses-action: ${{ inputs.uses-action }}
          runner: 'ubuntu-latest'
          create-pr: ${{ inputs.create-pr }}
          pr-branch: ${{ inputs.pr-branch }}
          pr-title: ${{ inputs.pr-title }}

  notify-failure:
    needs: [execute, execute-fallback]
    if: |
      always() &&
      (needs.execute.result == 'failure' || needs.execute.result == 'cancelled') &&
      (needs.execute-fallback.result == 'failure' || needs.execute-fallback.result == 'skipped')
    runs-on: ubuntu-latest
    permissions:
      issues: write
    steps:
      - name: Notify failure
        uses: mon2org/Common/.github/actions/core/notify@main

```

### 入力パラメータ仕様 (`inputs`)

| パラメータ名 | 必須 | 型 | デフォルト値 | 説明 |
| --- | --- | --- | --- | --- |
| `run-command` | No | `string` | `''` | 実行する CLI コマンド（`mise ...` で始まる必要あり） |
| `uses-action` | No | `string` | `''` | 実行する Action パス（`./` または `mon2org/` のみ許可） |
| `runner` | No | `string` | `'ubuntu-latest-1-vcpu'` | 実行ランナー |
| `create-pr` | No | `boolean` | `false` | タスク完了後に自動 PR を作成するかどうか |
| `pr-branch` | No | `string` | `'chore/sync-templates'` | 作成する PR ブランチ名 |
| `pr-title` | No | `string` | `'chore: テンプレート・共通設定の自動同期'` | 作成する PR タイトル |
| `timeout-minutes` | No | `number` | `15` | タイムアウト時間（分） |

### ジョブ制御フロー

```text
[ Start ]
    │
    ▼
[ Job: execute ] ──( 成功 )───────────────────────────────────► [ End: Success ]
    │
    ▼ ( 失敗かつ runner == 'ubuntu-latest-1-vcpu' )
[ Job: execute-fallback (ubuntu-latest) ] ──( 成功 )──────────► [ End: Success ]
    │
    ▼ ( 両ジョブともに失敗 / キャンセル )
[ Job: notify-failure ] (Issue 作成/更新) ─────────────────────► [ End: Failed ]

```

---

## 4. 構成アクション詳細

### ① `actions/core/execute` （パイプライン制御）

セキュリティガード、`mise` のセットアップ、コード取得、処理の実行、PR 作成までの一連のパイプラインを順次実行します。

* **処理切り分け:**
`run-command` が渡された場合は `run.sh` スクリプトを実行し、`uses-action` が渡された場合は指定のアクションを呼び出します。
* **エラー制御 (`run.sh`):**
* Exit Code `137` (OOM) および `126`/`127` (権限/コマンド不在) はリトライせず即時終了し、上位ワークフローのフォールバックへ渡します。
* その他の一時的エラーは最大2回（5秒間隔）リトライを行います。



### ② `actions/core/guard` （セキュリティ検証）

`checker.sh` により、以下のセキュリティルールを評価します。

| チェック項目 | 検証ルール | 違反時の動作 |
| --- | --- | --- |
| **パラメータ入力チェック** | `run-command` または `uses-action` のどちらか片方が必須 | ❌ エラー停止 |
| **External Fork 制限** | Fork リポジトリからの `pull_request` 実行を遮断 | ❌ エラー停止 |
| **PR 時の PR 作成禁止** | `pull_request` イベントでの `create-pr: true` を禁止 | ❌ エラー停止 |
| **コマンド構文** | `run-command` は `mise ` から始まる構文を強制 | ❌ エラー停止 |
| **アクション制限** | `uses-action` は `./` または `mon2org/` 開始を強制 | ❌ エラー停止 |
| **ランナー制限** | ホワイトリスト（`ubuntu-latest-1-vcpu`, `ubuntu-latest`, `macos-latest`, `windows-latest`）のみ許可 | ❌ エラー停止 |

### ③ `actions/core/notify` （障害通知）

`notify.sh` により、GitHub CLI (`gh`) を用いて同じタイトルの Open な Issue を検索します。

* 既存 Issue がある場合: ログ URL 付きコメントを投稿。
* 既存 Issue がない場合: `❌ ワークフロー失敗: <WORKFLOW_NAME>` のタイトルで `bug` ラベル付き Issue を新規作成。

---

## 5. 利用パターンと実装例

### パターン A: CLI コマンド（`run-command`）を動かす場合

`mise` タスクを呼び出して同期処理を行い、差分があれば PR を作成する標準パターンです。

```yaml
name: Sync Templates

on:
  schedule:
    - cron: '0 0 * * 1'
  workflow_dispatch:

jobs:
  sync:
    uses: mon2org/Common/.github/workflows/core.yml@main
    with:
      run-command: 'mise run sync-templates'
      create-pr: true
      pr-branch: 'chore/sync-templates'
      pr-title: 'chore: テンプレートの最新化'
    secrets: inherit

```

### パターン B: リポジトリ内の独自 Action（`uses-action`）を動かす場合

リポジトリ内に定義された Composite Action を共通基盤上で安全に実行するパターンです。

```yaml
name: Custom Action Execution

on:
  push:
    branches: [ main ]

jobs:
  run-custom:
    uses: mon2org/Common/.github/workflows/core.yml@main
    with:
      uses-action: './.github/actions/update/sync-templates'
      runner: 'ubuntu-latest'
    secrets: inherit

```