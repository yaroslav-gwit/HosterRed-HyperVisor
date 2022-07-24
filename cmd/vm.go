package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	jsonOutputVm bool

	vmCmd = &cobra.Command{
		Use:   "vm",
		Short: "VM related operations",
		Long:  `VM related operations, ie VM deloyment, stopping/starting the VMs, etc`,
		Run: func(cmd *cobra.Command, args []string) {
			main()
		},
	}
)

func main() {
	fmt.Println("test")
}
