// Package envutil は CLI が実行されている OS およびコンテナ/CI/仮想化環境を特定する機能を提供します。
// cmd/ 配下の各種 CLI ツールから共通して利用される内部パッケージです。
package envutil

import (
	"os"
	"runtime"
	"strings"
)

// EnvInfo は現在の実行環境に関する判定結果を保持する構造体です。
type EnvInfo struct {
	// OS は実行環境の OS 名（"darwin", "linux", "windows" 等）です。
	OS string

	// IsDarwin は macOS 環境で実行されているかどうかを示します。
	IsDarwin bool

	// IsLinux は Linux 環境で実行されているかどうかを示します。
	IsLinux bool

	// IsWindows は Windows 環境で実行されているかどうかを示します。
	IsWindows bool

	// IsWSL2 は Windows Subsystem for Linux (WSL/WSL2) 上で実行されているかどうかを示します。
	IsWSL2 bool

	// IsGitHubActions は GitHub Actions の CI 環境上で実行されているかどうかを示します。
	IsGitHubActions bool

	// IsCodespaces は GitHub Codespaces 上で実行されているかどうかを示します。
	IsCodespaces bool

	// IsDevContainer は VS Code DevContainer 等のコンテナ開発環境上で実行されているかどうかを示します。
	IsDevContainer bool

	// IsCI は CI/CD 環境全般（GitHub Actions, CircleCI 等）で実行されているかどうかを示します。
	IsCI bool
}

// Detect は現在の実行環境（OS、WSL、CI、コンテナ環境等）を自動検出して EnvInfo 構造体を返します。
func Detect() EnvInfo {
	goos := runtime.GOOS
	info := EnvInfo{
		OS:              goos,
		IsDarwin:        goos == "darwin",
		IsLinux:         goos == "linux",
		IsWindows:       goos == "windows",
		IsGitHubActions: os.Getenv("GITHUB_ACTIONS") == "true",
		IsCodespaces:    os.Getenv("CODESPACES") == "true",
		IsDevContainer:  os.Getenv("DEVCONTAINER") == "true" || os.Getenv("REMOTE_CONTAINERS") == "true",
		IsCI:            os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true",
	}

	if info.IsLinux {
		info.IsWSL2 = checkWSL()
	}

	return info
}

// checkWSL は /proc/version の内容または環境変数から WSL かどうかを判定します。
func checkWSL() bool {
	if os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSL_INTEROP") != "" {
		return true
	}

	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}

	versionStr := strings.ToLower(string(data))
	return strings.Contains(versionStr, "microsoft") || strings.Contains(versionStr, "wsl")
}
