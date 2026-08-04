# Core Notify (`actions/core/notify`)

`Core Notify` は、ワークフローの完全失敗時（メイン実行およびフォールバック実行がともに失敗した場合）に、通知およびトラッキング用 Issue を自動管理する Composite Action です。

---

## 📁 ファイル構成

```text
actions/core/notify/
├── action.yml    # Composite Action 定義
├── notify.sh     # 重複チェック＆ Issue 作成/コメント追記スクリプト
└── README.md     # 本ドキュメント

```

---

## 🔔 特徴・動作仕様

### 1. Issue の重複防止 (Idempotent Notification)

ワークフローが定期実行等で連続失敗した場合に Issue が大量乱立するのを防ぐため、`gh` CLI を使用して既存の Open な Issue を検索します。

* **同一タイトルの Open Issue が存在する場合:**
新規 Issue は作成せず、既存 Issue へ「実行ログへのリンク」を含めた追加コメントを投稿します。
* **存在しない場合:**
タイトル `❌ ワークフロー失敗: <WORKFLOW_NAME>` で新規 Issue を作成し、`bug` ラベルを付与します。

### 2. 暗黙的なコンテキスト参照

`action.yml` 側で GitHub Actions の標準環境変数（`${{ github.workflow }}`, `${{ github.repository }}`, `${{ github.run_id }}` 等）を収集して `notify.sh` に渡すため、呼び出し側でパラメータ指定を行う必要はありません。

---

## 🔑 必要な権限 (Permissions)

このアクションを実行するジョブには、Issue の検索・作成・コメント追加を行うための書き込み権限が必要です。

```yaml
permissions:
  issues: write

```

---

## 🚀 使用方法 (Usage)

ワークフロー（`core.yml` 等）の最終通知ジョブ（`notify-failure`）から呼び出して利用します。

```yaml
jobs:
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