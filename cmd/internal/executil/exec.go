// Package executil は外部コマンドの存在確認および実行に関する汎用ユーティリティを提供します。
// cmd/ 配下の各種 CLI ツールから共通して利用される内部パッケージです。
package executil

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
)

// IsAvailable は指定されたコマンドがシステムの PATH 上に存在し、実行可能かどうかを確認します。
// 実行可能な場合は true、存在しないか権限がない場合は false を返します。
func IsAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// Run は標準出力および標準エラーを現在のプロセスに接続して外部コマンドを実行します（非対話的実行）。
// インストールコマンドやスクリプトなどのバックグラウンド/バッチ処理に適しています。
func Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunQuiet は標準出力および標準エラーを破棄して外部コマンドを実行します。
// 存在確認やサイレントな事前判定に適しています。
func RunQuiet(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

// RunInteractive は標準入力、標準出力、標準エラーをすべて現在のプロセスに接続して外部コマンドを実行します。
func RunInteractive(ctx context.Context, name string, args []string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// IsExitError はエラーがコマンド実行後の終了ステータスエラー (*exec.ExitError) かどうかを判定します。
func IsExitError(err error) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr)
}

// GetExitCode はエラーから終了ステータスコードを取得します。
func GetExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}