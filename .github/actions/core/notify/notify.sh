#!/usr/bin/env bash
set -euo pipefail

ISSUE_TITLE="❌ ワークフロー失敗: ${WORKFLOW_NAME:-Workflow}"
WORKFLOW_URL="${SERVER_URL:-https://github.com}/${REPOSITORY:-}/actions/runs/${RUN_ID:-}"

echo "🔔 失敗通知処理を開始します..."

# 既存の Open な Issue を検索 (重複防止)
EXISTING_ISSUE_NUM=$(gh issue list \
  --repo "${REPOSITORY}" \
  --state open \
  --search "$ISSUE_TITLE in:title" \
  --json number \
  --jq '.[0].number' 2>/dev/null || echo "")

if [ -n "$EXISTING_ISSUE_NUM" ] && [ "$EXISTING_ISSUE_NUM" != "null" ]; then
  echo "⚠️ 既存の Issue (#$EXISTING_ISSUE_NUM) に失敗ログを追加します。"
  gh issue comment "$EXISTING_ISSUE_NUM" \
    --repo "${REPOSITORY}" \
    --body "🔄 再度失敗しました。[実行ログを確認する]($WORKFLOW_URL)"
else
  echo "🆕 新しい失敗通知 Issue を作成します。"
  gh issue create \
    --repo "${REPOSITORY}" \
    --title "$ISSUE_TITLE" \
    --body "自動処理に失敗しました（フォールバック実行含む）。[実行ログを確認する]($WORKFLOW_URL)" \
    --label "bug" || echo "⚠️ Issue の作成に失敗しました（権限不足の可能性があります）。"
fi