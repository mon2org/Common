#!/usr/bin/env bash
set -euo pipefail

echo "🔒 Security Guard: 実行安全性を検証中..."

log_error_to_summary() {
  local message="$1"
  if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
    {
      echo "### ❌ Security Guard Error"
      echo "$message"
    } >> "$GITHUB_STEP_SUMMARY"
  fi
}

# 1. 実行対象の指定チェック（どちらか必須）
if [ -z "${RUN_COMMAND:-}" ] && [ -z "${USES_ACTION:-}" ]; then
  msg="'run-command' または 'uses-action' のいずれか一方を指定する必要があります。"
  echo "❌ [SECURITY ERROR] $msg"
  log_error_to_summary "$msg"
  exit 1
fi

if [ "${EVENT_NAME:-}" = "pull_request" ] && [ "${IS_FORK:-}" = "true" ]; then
  msg="Fork リポジトリからの Pull Request 実行は許可されていません。"
  echo "❌ [SECURITY ERROR] $msg"
  log_error_to_summary "$msg"
  exit 1
fi

if [ "${EVENT_NAME:-}" = "pull_request" ] && [ "${CREATE_PR:-}" = "true" ]; then
  msg="PR イベントから 'create-pr: true' を指定することは禁止されています。"
  echo "❌ [SECURITY ERROR] $msg"
  log_error_to_summary "$msg"
  exit 1
fi

# 2. Command 指定時の構文検証
if [ -n "${RUN_COMMAND:-}" ] && [[ ! "${RUN_COMMAND:-}" =~ ^mise[[:space:]] ]]; then
  msg="実行コマンドは 'mise ...' で始まる必要があります。受領: \`${RUN_COMMAND:-}\`"
  echo "❌ [SECURITY ERROR] $msg"
  log_error_to_summary "$msg"
  exit 1
fi

# 3. Action 指定時のホワイトリスト検証 (ローカルパス ./.github または mon2org 組織のみ許可)
if [ -n "${USES_ACTION:-}" ] && [[ ! "${USES_ACTION:-}" =~ ^(\./|mon2org/) ]]; then
  msg="許可されていない Action が指定されました: \`${USES_ACTION:-}\` (自リポジトリ './' または 'mon2org/' 組織のみ許可)"
  echo "❌ [SECURITY ERROR] $msg"
  log_error_to_summary "$msg"
  exit 1
fi

ALLOWED_RUNNERS=("ubuntu-latest-1-vcpu" "ubuntu-latest" "macos-latest" "windows-latest")
RUNNER_ALLOWED=false
for allowed in "${ALLOWED_RUNNERS[@]}"; do
  if [ "${RUNNER:-}" = "$allowed" ]; then
    RUNNER_ALLOWED=true
    break
  fi
done

if [ "$RUNNER_ALLOWED" = "false" ]; then
  msg="不許可のランナーが指定されました: \`${RUNNER:-}\`"
  echo "❌ [SECURITY ERROR] $msg"
  log_error_to_summary "$msg"
  exit 1
fi

if [ "${EVENT_NAME:-}" = "pull_request" ]; then
  CAN_PUSH=$(gh api "repos/${GITHUB_REPOSITORY:-}" --jq '.permissions.push' 2>/dev/null || echo "false")
  if [ "$CAN_PUSH" = "true" ]; then
    echo "⚠️ [SECURITY WARNING] PR イベントですが書き込み権限が検知されました。"
  fi
fi

echo "✅ Security Guard: すべての検証をパスしました。"