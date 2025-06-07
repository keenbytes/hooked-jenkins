package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/keenbytes/broccli/v3"
	"github.com/keenbytes/octo-linter/pkg/loglevel"
)

func main() {
	cli := broccli.NewBroccli("hooked-jenkins", "Tiny API to receive GitHub Webhooks and trigger Jenkins jobs", "mg@keenbytes.co")

	cmd := cli.Command("start", "Start API", startHandler)
	cmd.Flag("config", "c", "FILE", "Configuration file", broccli.TypePathFile, broccli.IsRegularFile|broccli.IsExistent)
	cmd.Flag("loglevel", "l", "", "One of NONE,ERR,WARN,DEBUG", broccli.TypeString, 0)

	_ = cli.Command("version", "Prints version", versionHandler)
	if len(os.Args) == 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		os.Args = []string{"App", "version"}
	}

	os.Exit(cli.Run(context.Background()))
}

func versionHandler(ctx context.Context, c *broccli.Broccli) int {
	fmt.Fprintf(os.Stdout, VERSION+"\n")
	return 0
}

func startHandler(ctx context.Context, c *broccli.Broccli) int {
	logLevel := loglevel.GetLogLevelFromString(c.Flag("loglevel"))

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	app := &hookedJenkins{}

	cfg := &config{}
	err := cfg.readFile(c.Flag("config"))
	if err != nil {
		slog.Error(fmt.Sprintf("error reading config file: %s", err.Error()))
		return 31
	}

	app.config = cfg

	done := make(chan bool)
	go app.startAPI()
	<-done

	return 0
}
