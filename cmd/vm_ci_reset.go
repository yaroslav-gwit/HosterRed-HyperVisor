package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	newVmName string

	vmCiResetCmd = &cobra.Command{
		Use:   "cireset",
		Short: "Reset VM's passwords, ssh keys, and network config (useful after VM migration)",
		Long:  `Reset VM's passwords, ssh keys, and network config (useful after VM migration)`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(args[0])
			fmt.Println(newVmName)
		},
	}
)
