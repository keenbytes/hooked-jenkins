package main

import (
	"context"
	"fmt"
	"os"

	"gopkg.pl/mikogs/octo-linter/pkg/loglevel"
	"gopkg.pl/mikogs/broccli/v3"
)

func main() {
	cli := broccli.NewBroccli("hooked-jenkins", "Tiny API to receive GitHub Webhooks and trigger Jenkins jobs", "mg@computerclub.pl")

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

	app := &hookedJenkins{
		logLevel: logLevel,
	}

	cfg := &config{
		logLevel: logLevel,
	}
	err := cfg.readFile(c.Flag("config"))
	if err != nil {
		printErr(logLevel, err, "error reading config file")
		return 31
	}

	app.config = cfg

	done := make(chan bool)
	go app.startAPI()
	<-done

	return 0
}

func printErr(logLevel int, err error, msg string) {
	if logLevel == loglevel.LogLevelNone {
		return
	}

	fmt.Fprintf(os.Stderr, "!!!:%s: %s\n", msg, err.Error())
}
