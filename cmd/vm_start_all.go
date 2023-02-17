package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var (
	vmStartAllCmd = &cobra.Command{
		Use:   "start-all",
		Short: "Start all VMs deployed on this system",
		Long:  `Start all VMs deployed on this system`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			vmStartAll()
		},
	}
)

func vmStartAll() {
	sleepTime := 5
	for _, vm := range getAllVms() {
		vmConfigVar := vmConfig(vm)
		if vmConfigVar.ParentHost != GetHostName() {
			continue
		} else if vmConfigVar.LiveStatus == "production" || vmConfigVar.LiveStatus == "prod" {
			vmStart(vm)
			time.Sleep(time.Second * time.Duration(sleepTime))
			if sleepTime < 30 {
				sleepTime = sleepTime + 1
			}
		} else {
			continue
		}
	}
}
