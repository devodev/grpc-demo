package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	loggerOutput  = os.Stderr
	defaultOutput = os.Stdout
)

// Execute executes the root command.
func Execute() {
	rootCmd := newCommandRoot()
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
	}
}

func writeOut(line string) {
	fmt.Fprintln(defaultOutput, line)
}

func newCommandRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "client",
		Short:   "Interact with the Hub.",
		Version: "0.1.0",
	}
	cmd.AddCommand(
		newCommandFluentd(),
		newCommandHub(),
		newCommandSystemd(),
	)
	return cmd
}
