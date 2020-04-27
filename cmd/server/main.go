package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	loggerOutput  = os.Stderr
	defaultOutput = os.Stdout
)

// Execute executes the root command.
func Execute() error {
	rootCmd := newCommandRoot()
	return rootCmd.Execute()
}

func writeOut(line string) {
	fmt.Fprintln(defaultOutput, line)
}

func newCommandRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Short:   "gRPC server.",
		Version: "0.1.0",
	}
	cmd.AddCommand(
		newCommandServe(),
	)
	return cmd
}

func main() {
	if err := Execute(); err != nil {
		writeOut(err.Error())
		os.Exit(1)
	}
}
