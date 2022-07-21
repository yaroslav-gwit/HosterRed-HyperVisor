package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	address string

	hostCmd = &cobra.Command{
		Use:   "host",
		Short: "Host related operations",
		Long:  `Host related operations, ie set host name, get basic host info, etc`,
		Run: func(cmd *cobra.Command, args []string) {
			main()
		},
	}
)

func main() {
	fmt.Println(freeRam())
}

func freeRam() string {
	var vFreeCount string
	var vFreeCountArg1 = "sysctl"
	var vFreeCountArg2 = "-nq"
	var vFreeCountArg3 = "vm.stats.vm.v_free_count"

	var cmd = exec.Command(vFreeCountArg1, vFreeCountArg2, vFreeCountArg3)
	var stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func freeRam/vFreeCount: There has been an error:", err)
		os.Exit(1)
	} else {
		vFreeCount = string(stdout)
	}

	var hwPagesize string
	var hwPagesizeArg1 = "sysctl"
	var hwPagesizeArg2 = "-nq"
	var hwPagesizeArg3 = "hw.pagesize"
	cmd = exec.Command(hwPagesizeArg1, hwPagesizeArg2, hwPagesizeArg3)
	stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func freeRam/hwPagesize: There has been an error:", err)
		os.Exit(1)
	} else {
		hwPagesize = string(stdout)
	}

	fmt.Println(vFreeCount)
	fmt.Println(hwPagesize)
	var vFreeCountInt, _ = strconv.Atoi(vFreeCount)
	var hwPagesizeInt, _ = strconv.Atoi(hwPagesize)
	fmt.Println(vFreeCountInt)
	fmt.Println(hwPagesize)

	var finalResult = vFreeCountInt * hwPagesizeInt

	return strconv.Itoa(finalResult)
}
