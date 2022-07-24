package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	jsonOutputVm bool

	vmCmd = &cobra.Command{
		Use:   "host",
		Short: "Host related operations",
		Long:  `Host related operations, ie set host name, get basic host info, etc`,
		Run: func(cmd *cobra.Command, args []string) {
			main()
		},
	}
)

func main() {
	fmt.Println("test")
}
