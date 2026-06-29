package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	cliFlagDBDSN    = "db-dsn"
	cliFlagDBDir    = "db-dir"
	cliFlagLogLevel = "log-level"
	cliFlagLogDevel = "log-devel"
)

func bindDBDSNFlag(fs *pflag.FlagSet) {
	fs.StringSlice(cliFlagDBDSN, []string{}, "Database data source name to use (can be specified multiple times, format: [name=]path).")
}

func bindDBDirFlag(fs *pflag.FlagSet) {
	fs.String(cliFlagDBDir, "", "Directory containing database files (.db, .sqlite, .sqlite3).")
}

func createMainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "sqlite-rest",
		Short:        "Serve a RESTful API from a SQLite database",
		SilenceUsage: true,
	}

	cmd.PersistentFlags().
		Int8(cliFlagLogLevel, 5, "Log level to use. Use 8 or more for verbose log.")
	cmd.PersistentFlags().
		Bool(cliFlagLogDevel, false, "Enable devel log format?")

	cmd.AddCommand(
		createServeCmd(),
		createMigrateCmd(),
	)

	cmd.CompletionOptions.DisableDefaultCmd = true

	return cmd
}

func main() {
	cmd := createMainCmd()

	if cmd.Execute() != nil {
		os.Exit(1)
	}
}
