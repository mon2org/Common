// Package cliutil は CLI ツールのバージョン情報管理や Usage（ヘルプ）表示の標準化機能を提供します。
// cmd/ 配下の各種 CLI ツール間で統一されたインターフェースと出力形式を実現するための内部パッケージです。
package cliutil

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

// ErrVersionPrinted はバージョン情報が出力され、プログラムを正常終了（exit 0）させるべきであることを示すセンチネルエラーです。
var ErrVersionPrinted = errors.New("version printed")

// BuildInfo は CLI ツールのバージョンおよびビルド時のメタデータを保持する構造体です。
type BuildInfo struct {
	// Version はセマンティックバージョン（例: "1.0.0" や "v0.1.0"）です。
	Version string

	// Commit は Git の短縮コミットハッシュ（例: "a1b2c3d"）です。
	Commit string

	// Date はビルド日時（またはコミット日時）です。
	Date string

	// Dirty は未コミットの変更が含まれた状態でビルドされたかどうかを示します。
	Dirty bool
}

// GetBuildInfo は指定されたデフォルト値（ldflags経由など）を優先しつつ、
// 未指定の場合は runtime/debug.ReadBuildInfo() から動的に Git 情報を自動取得して返します。
func GetBuildInfo(defaultVersion, defaultCommit, defaultDate string) BuildInfo {
	info := BuildInfo{
		Version: defaultVersion,
		Commit:  defaultCommit,
		Date:    defaultDate,
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		if info.Version == "" {
			info.Version = "dev"
		}
		return info
	}

	// 1. バージョン文字列の動的補完
	if info.Version == "" || info.Version == "dev" {
		if buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
			info.Version = buildInfo.Main.Version
		} else {
			info.Version = "dev"
		}
	}

	// 2. VCS (Git) メタデータの自動解析
	var vcsTime string
	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs.revision":
			if info.Commit == "" {
				if len(setting.Value) >= 7 {
					info.Commit = setting.Value[:7]
				} else {
					info.Commit = setting.Value
				}
			}
		case "vcs.time":
			vcsTime = setting.Value
		case "vcs.modified":
			if setting.Value == "true" {
				info.Dirty = true
			}
		}
	}

	// 3. 日時の整形
	if info.Date == "" && vcsTime != "" {
		if t, err := time.Parse(time.RFC3339, vcsTime); err == nil {
			info.Date = t.Local().Format("2006-01-02 15:04:05")
		} else {
			info.Date = vcsTime
		}
	}

	return info
}

// String は BuildInfo を人間が読みやすいバージョン文字列にフォーマットして返します。
func (b BuildInfo) String(appName string) string {
	v := b.Version
	if v == "" {
		v = "dev"
	}
	if b.Commit != "" {
		dirtyMarker := ""
		if b.Dirty {
			dirtyMarker = "-dirty"
		}
		v = fmt.Sprintf("%s (%s%s)", v, b.Commit, dirtyMarker)
	}
	if b.Date != "" {
		v = fmt.Sprintf("%s built at %s", v, b.Date)
	}
	return fmt.Sprintf("%s version %s", appName, v)
}

// AppMeta は CLI ツールのヘルプ（Usage）表示に必要なメタデータを保持する構造体です。
type AppMeta struct {
	// Name はアプリケーションの名前です。
	Name string

	// Description はアプリケーションの概要説明です。
	Description string

	// Usage は使用方法の構文パターンです。
	Usage string

	// Examples は使用例の文字列リストです。
	Examples []string
}

// SetupUsage は指定された FlagSet に対して統一フォーマットの Usage 出力関数を設定します。
func SetupUsage(fs *flag.FlagSet, meta AppMeta) {
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s - %s\n\n", meta.Name, meta.Description)
		if meta.Usage != "" {
			fmt.Fprintf(os.Stderr, "使用方法:\n  %s\n\n", meta.Usage)
		}
		if len(meta.Examples) > 0 {
			fmt.Fprintf(os.Stderr, "例:\n")
			for _, ex := range meta.Examples {
				fmt.Fprintf(os.Stderr, "  %s\n", ex)
			}
			fmt.Fprintln(os.Stderr)
		}
		fmt.Fprintf(os.Stderr, "オプション:\n")
		fs.PrintDefaults()
	}
}
