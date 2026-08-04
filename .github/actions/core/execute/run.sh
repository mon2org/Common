#!/usr/bin/env bash
set -euo pipefail

if [ -z "${RUN_COMMAND:-}" ]; then
  echo "❌ [ERROR] RUN_COMMAND 環境変数が設定されていません。"
  exit 1
fi

MAX_RETRIES=2
RETRY_DELAY_SECS=5
ATTEMPT=1

while [ $ATTEMPT -le $MAX_RETRIES ]; do
  echo "🚀 タスクを実行します (試行 $ATTEMPT/$MAX_RETRIES): $RUN_COMMAND"

  set +e
  eval "$RUN_COMMAND"
  EXIT_CODE=$?
  set -e

  if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ タスクが正常に完了しました。"
    exit 0
  fi

  echo "⚠️ 試行 $ATTEMPT が失敗しました (Exit Code: $EXIT_CODE)"

  case $EXIT_CODE in
    137)
      echo "⚠️ [OOM DETECTED] メモリ不足を検知しました。フォールバック起動のため即時終了します。"
      break
      ;;
    126|127)
      echo "❌ [FATAL ERROR] コマンドの実行権限がないか、指定された設定が存在しません。即時終了します。"
      break
      ;;
    *)
      if [ $ATTEMPT -lt $MAX_RETRIES ]; then
        echo "🔄 一時的なエラーの可能性があるため、${RETRY_DELAY_SECS}秒後にリトライします..."
        sleep $RETRY_DELAY_SECS
      fi
      ;;
  esac

  ATTEMPT=$((ATTEMPT + 1))
done

if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
  {
    echo "### ❌ Main Task Failed"
    echo "- **Executed Command:** \`$RUN_COMMAND\`"
    echo "- **Exit Code:** \`$EXIT_CODE\`"

    case $EXIT_CODE in
      137)
        echo "- **Cause:** ⚠️ **Out of Memory (OOM)** — メモリ不足により強制終了されました。"
        ;;
      127)
        echo "- **Cause:** ⚠️ **Command Not Found** — 指定された mise タスクまたはコマンドが存在しません。"
        ;;
      126)
        echo "- **Cause:** ⚠️ **Permission Denied** — コマンドの実行権限がありません。"
        ;;
      *)
        echo "- **Cause:** ⚠️ タスクの実行に失敗しました（最大リトライ数超過）。"
        ;;
    esac
  } >> "$GITHUB_STEP_SUMMARY"
fi

exit "$EXIT_CODE"