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
	// Website related checks
	rootCmd.AddCommand(hostCmd)

	// Address flag
	// webCmd.Flags().StringVarP(&address, "address", "a", "", "Website address (required)")
	// webCmd.Flags().StringVarP(&string_present, "string", "s", "", "Check if this string exists on the page")
	// webCmd.MarkFlagsRequiredTogether("address", "string")

	// Output JSON to the screen
	// webCmd.Flags().BoolVar(&json_output, "json", false, "Use JSON output instead of table output. Useful for passing to other programs")
	// webCmd.Flags().BoolVar(&save_results, "save-results", false, "Save JSON results to a file")
	// webCmd.Flags().StringVar(&results_file, "results-file", "/tmp/results.json", "Optionally set the location of the file to save the results to")

	// Port flag
	// webCmd.Flags().StringVar(&port, "port", "443", "Website port")
	// webCmd.Flags().StringVar(&protocol, "protocol", "https", "Website connection protocol")

	// Page to check
	// webCmd.Flags().StringVar(&pageToCheck, "page", "/", "Endpoint webpage to check")

	// File flag
	// webCmd.Flags().StringVarP(&file_database, "file", "f", "db.json", "Use JSON file database to check multiple servers at once")

	// Mutually exculusive flags
	// webCmd.MarkFlagsMutuallyExclusive("file", "address")
	// webCmd.MarkFlagsMutuallyExclusive("file", "port")
	// webCmd.MarkFlagsMutuallyExclusive("file", "protocol")
	// webCmd.MarkFlagsMutuallyExclusive("file", "string")

	// webCmd.MarkFlagsMutuallyExclusive("save-results", "address")
	// webCmd.MarkFlagsMutuallyExclusive("save-results", "port")
	// webCmd.MarkFlagsMutuallyExclusive("save-results", "protocol")
	// webCmd.MarkFlagsMutuallyExclusive("save-results", "string")

	// webCmd.MarkFlagsMutuallyExclusive("results-file", "address")
	// webCmd.MarkFlagsMutuallyExclusive("results-file", "port")
	// webCmd.MarkFlagsMutuallyExclusive("results-file", "protocol")
	// webCmd.MarkFlagsMutuallyExclusive("results-file", "string")

	// Print version
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of HosterRed",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("HosterRed: v0.1")
	},
}
