# Core Guard (`actions/core/guard`)

`Core Guard` は、共通ワークフロー（`core.yml` および `core/execute`）の実行前処理として呼び出されるセキュリティ検証 Composite Action です。

Poisoned Pipeline Execution (PPE) 攻撃や設定ミスによる不正利用を未然に防ぎ、CI/CD 実行基盤の安全性を担保します。

---

## 📁 ファイル構成

```text
actions/core/guard/
├── action.yml    # Composite Action 定義（パラメータ受領・スクリプト呼び出し）
├── checker.sh    # 検証ロジック本体（Bash スクリプト）
└── README.md     # 本ドキュメント
```

---

## 🛡️ 検証ルール (Guard Checks)

`checker.sh` 内で以下の 5 つの検証を順次実行し、不適切なリクエストを数秒で即時ブロック（Fail-Fast）します。

| # | チェック項目 | 説明 | 判定エラー時の動作 |
| --- | --- | --- | --- |
| **1** | **External Fork 判定** | 外部 Fork リポジトリからの Pull Request 実行を検知 | ❌ エラー停止 |
| **2** | **PR 時の PR 作成禁止** | `pull_request` イベント時に `create-pr: true` が渡されていないか確認 | ❌ エラー停止 |
| **3** | **コマンド構文チェック** | 実行コマンド（`run-command`）が `mise ` で始まっているか強制 | ❌ エラー停止 |
| **4** | **ランナー制限** | `runner` が組織のホワイトリストに含まれているか確認 | ❌ エラー停止 |
| **5** | **過剰権限の警告** | `pull_request` 実行時に書き込み権限（`contents: write`）がないか照会 | ⚠️ 警告ログ出力 |

### 💡 許可されているランナー（ホワイトリスト）

* `ubuntu-latest-1-vcpu`
* `ubuntu-latest`
* `macos-latest`
* `windows-latest`

---

## 📥 入力パラメータ (Inputs)

| パラメータ | 必須 | デフォルト値 | 説明 |
| --- | --- | --- | --- |
| `run-command` | **Yes** | - | 実行予定のコマンド（`mise ...` で始まる必要あり） |
| `runner` | **Yes** | - | 実行予定のランナー名 |
| `create-pr` | **Yes** | - | PR 作成フラグ（`'true'` または `'false'`） |

---

## 🚀 使用方法 (Usage)

基本的には `actions/core/execute` 内の先頭ステップから自動的に呼び出されます。単体で呼び出す場合は以下のように記述します。

```yaml
- name: Run Security Guard
  uses: mon2org/Common/.github/actions/core/guard@main
  with:
    run-command: 'mise run test'
    runner: 'ubuntu-latest-1-vcpu'
    create-pr: 'false'

```

---

## 🚨 エラー時の挙動

いずれかの検証でエラーが検出された場合、スクリプトは `exit 1` で終了し、後続のステップ（コードのチェックアウトや環境セットアップ等）は実行されません。

また、失敗の理由は GitHub Actions の **Job Summary（`$GITHUB_STEP_SUMMARY`）** 画面に自動で書き出されるため、実行ログを検索することなく修正箇所を把握できます。