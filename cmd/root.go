package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hoster",
	Short: "HosterCore is a highly opinionated Bhyve automation platform written in Go",

	Run: func(cmd *cobra.Command, args []string) {
		HostMain()
		VmListMain()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Host command section
	rootCmd.AddCommand(hostCmd)
	hostCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON (useful for automation)")

	// VM command section
	rootCmd.AddCommand(vmCmd)
	vmCmd.AddCommand(vmListCmd)
	vmListCmd.Flags().BoolVarP(&jsonOutputVm, "json", "j", false, "Output as JSON (useful for automation)")
	vmListCmd.Flags().BoolVarP(&jsonPrettyOutputVm, "json-pretty", "", false, "Pretty JSON Output")
	vmListCmd.Flags().BoolVarP(&tableUnixOutputVm, "unix-style", "u", false, "Show Unix style table (useful for bash scripting)")

	// API command section
	rootCmd.AddCommand(apiCmd)
	apiCmd.Flags().IntVarP(&apiServerPort, "port", "p", 3000, "Specify the port to listen on")
	apiCmd.Flags().StringVarP(&apiServerUser, "user", "u", "admin", "Username for API authentication")
	apiCmd.Flags().StringVarP(&apiServerPassword, "password", "", "123456", "Password for API authentication")

	// Version command section
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of HosterCore",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("HosterRed v0.1, Golang version")
	},
}
