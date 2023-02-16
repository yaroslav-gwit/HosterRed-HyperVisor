package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	vmStartCmd = &cobra.Command{
		Use:   "start [VM name]",
		Short: "Start a particular VM using it's name",
		Long:  `Start a particular VM using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			fmt.Println(args[0])
		},
	}
)
