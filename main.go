package main

import (
	"fmt"
	"os"

	"gopkg.pl/mikogs/octo-linter/pkg/loglevel"
	"gopkg.pl/phings/broccli"
)

func main() {
	cli := broccli.NewCLI("hooked-jenkins", "Tiny API to receive GitHub Webhooks and trigger Jenkins jobs", "mg@computerclub.pl")

	cmd := cli.AddCmd("start", "Start API", startHandler)
	cmd.AddFlag("config", "c", "FILE", "Configuration file", broccli.TypePathFile, broccli.IsRegularFile|broccli.IsExistent)
	cmd.AddFlag("loglevel", "l", "", "One of NONE,ERR,WARN,DEBUG", broccli.TypeString, 0)

	_ = cli.AddCmd("version", "Prints version", versionHandler)
	if len(os.Args) == 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		os.Args = []string{"App", "version"}
	}

	os.Exit(cli.Run())
}

func versionHandler(c *broccli.CLI) int {
	fmt.Fprintf(os.Stdout, VERSION+"\n")
	return 0
}

func startHandler(c *broccli.CLI) int {
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
