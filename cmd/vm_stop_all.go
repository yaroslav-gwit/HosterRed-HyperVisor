package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var (
	vmStopAllCmd = &cobra.Command{
		Use:   "stop-all",
		Short: "Stop all VMs deployed on this system",
		Long:  `Stop all VMs deployed on this system`,
		Run: func(cmd *cobra.Command, args []string) {
			vmStopAll()
		},
	}
)

func vmStopAll() {
	sleepTime := 3
	for _, vm := range getAllVms() {
		vmStop(vm)
		time.Sleep(time.Second * time.Duration(sleepTime))
	}
}
