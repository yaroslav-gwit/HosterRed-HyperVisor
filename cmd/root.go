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
	hostCmd.Flags().BoolVarP(&jsonHostInfoOutput, "json", "j", false, "Output as JSON (useful for automation)")
	hostCmd.Flags().BoolVarP(&jsonPrettyHostInfoOutput, "json-pretty", "", false, "Pretty JSON Output")

	// VM command section
	rootCmd.AddCommand(vmCmd)

	// VM cmd -> list
	vmCmd.AddCommand(vmListCmd)
	vmListCmd.Flags().BoolVarP(&jsonOutputVm, "json", "j", false, "Output as JSON (useful for automation)")
	vmListCmd.Flags().BoolVarP(&jsonPrettyOutputVm, "json-pretty", "", false, "Pretty JSON Output")
	vmListCmd.Flags().BoolVarP(&tableUnixOutputVm, "unix-style", "u", false, "Show Unix style table (useful for bash scripting)")

	// VM cmd -> info
	vmCmd.AddCommand(vmInfoCmd)
	vmInfoCmd.Flags().BoolVarP(&jsonVmInfo, "json", "j", false, "Output as JSON (useful for automation)")
	vmInfoCmd.Flags().BoolVarP(&jsonPrettyVmInfo, "json-pretty", "", false, "Pretty JSON Output")

	// VM cmd -> start
	vmCmd.AddCommand(vmStartCmd)

	// VM cmd -> start all
	vmCmd.AddCommand(vmStartAllCmd)

	// VM cmd -> stop
	vmCmd.AddCommand(vmStopCmd)

	// VM cmd -> stop all
	vmCmd.AddCommand(vmStopAllCmd)

	// VM cmd -> snapshot
	vmCmd.AddCommand(vmZfsSnapshotCmd)
	vmZfsSnapshotCmd.Flags().StringVarP(&snapshotType, "stype", "t", "custom", "Snapshot type")
	vmZfsSnapshotCmd.Flags().IntVarP(&snapshotsToKeep, "keep", "k", 5, "Number of snapshots to keep for this specific snapshot type")

	// VM cmd -> snapshot all
	vmCmd.AddCommand(vmZfsSnapshotAllCmd)
	vmZfsSnapshotAllCmd.Flags().StringVarP(&snapshotAllType, "stype", "t", "custom", "Snapshot type")
	vmZfsSnapshotAllCmd.Flags().IntVarP(&snapshotsAllToKeep, "keep", "k", 5, "Number of snapshots to keep for this specific snapshot type")

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
		fmt.Println("HosterCore v0.1, version based on Golang")
	},
}
