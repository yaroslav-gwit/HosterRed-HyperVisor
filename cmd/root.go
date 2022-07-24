package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hoster",
	Short: "HosterRed is a highly opinionated Bhyve automation library written in Go",

	Run: func(cmd *cobra.Command, args []string) {
		// Empty function
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(hostCmd)
	rootCmd.AddCommand(vmCmd)
	hostCmd.Flags().BoolVarP(&jsonOutput, "json-output", "j", false, "Output as JSON (useful for automation)")
	vmCmd.Flags().BoolVarP(&jsonOutputVm, "json-output", "j", false, "Output as JSON (useful for automation)")

	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of HosterRed",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("HosterRed: v0.1")
	},
}
