package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/webdevmeg42/go-figure/internal/config"
	"github.com/webdevmeg42/go-figure/internal/reporter"
	"github.com/webdevmeg42/go-figure/internal/runner"
)

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "go-figure",
	Short: "A lightweight HTTP test runner for CI pipelines",
}

var runCmd = &cobra.Command{
	Use:   "run <file>",
	Short: "Run HTTP tests from a YAML file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file := args[0]
		fmt.Fprintf(os.Stdout, "\ngo-figure v%s  ·  %s\n", version, file)

		suite, err := config.Parse(file)
		if err != nil {
			return fmt.Errorf("loading %s: %w", file, err)
		}

		results := runner.Run(suite)
		if !reporter.Print(os.Stdout, results) {
			os.Exit(1)
		}
		return nil
	},
}

// Execute wires the command tree and runs cobra.
func Execute() {
	rootCmd.AddCommand(runCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
