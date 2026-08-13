package main

import (
	"flag"
	"fmt"

	"github.com/mon2org/Common/cmd/internal/cliutil"
)

var (
	Version = ""
	Commit  = ""
	Date    = ""
)

type Config struct {
	Args []string
}

func parseFlags() (*Config, error) {
	cfg := &Config{}
	var showVersion bool

	meta := cliutil.AppMeta{
		Name:        "template",
		Description: "CLI ツールのテンプレート",
		Usage:       "template [オプション]",
	}
	cliutil.SetupUsage(flag.CommandLine, meta)

	flag.BoolVar(&showVersion, "version", false, "バージョン情報を表示")
	flag.BoolVar(&showVersion, "v", false, "バージョン情報を表示 (短縮形)")

	flag.Parse()

	if showVersion {
		buildInfo := cliutil.GetBuildInfo(Version, Commit, Date)
		fmt.Println(buildInfo.String("template"))
		return nil, cliutil.ErrVersionPrinted
	}

	cfg.Args = flag.Args()
	return cfg, nil
}