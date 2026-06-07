package config

import (
	"flag"
	"os"
	"testing"
)

func resetFlags(t *testing.T) {
	t.Helper()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flagLogLevel = flag.String("log-level", "", "log level")
	flagLogOutput = flag.String("log-output", "", "log output")
	flagLogFile = flag.String("log-file", "", "log file")
	flagDB = flag.String("db", "", "db path")
}

func TestParseFlags(t *testing.T) {
	resetFlags(t)
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"data-nexus", "--log-level=debug", "--log-output=file", "--log-file=/tmp/app.log", "--db=/data/app.db"}

	cli := ParseFlags()
	if cli.LogLevel != "debug" || cli.LogOutput != "file" || cli.LogFile != "/tmp/app.log" || cli.DBPath != "/data/app.db" {
		t.Fatalf("unexpected cli overrides: %+v", cli)
	}
}

func TestParseFlagsPositionalDatabaseArg(t *testing.T) {
	resetFlags(t)
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"data-nexus", "--log-level=info", "/Users/me/project.sqlite3"}

	cli := ParseFlags()
	if cli.DBPath != "/Users/me/project.sqlite3" {
		t.Fatalf("expected positional db path, got %q", cli.DBPath)
	}
}

func TestFirstDatabaseArg(t *testing.T) {
	if got := firstDatabaseArg([]string{"-v", "notes.db", "other.txt"}); got != "notes.db" {
		t.Fatalf("got %q want notes.db", got)
	}
	if got := firstDatabaseArg([]string{"-v", "other.txt"}); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}
