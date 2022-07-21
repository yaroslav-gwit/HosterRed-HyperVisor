package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

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
	// GET SYSCTL "vm.stats.vm.v_free_count" AND RETURN THE VALUE
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

	var vFreeCountList []string
	for _, i := range strings.Split(vFreeCount, "\n") {
		if len(i) > 1 {
			vFreeCountList = append(vFreeCountList, i)
		}
	}
	vFreeCount = vFreeCountList[0]

	// GET SYSCTL "hw.pagesize" AND RETURN THE VALUE
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
	var hwPagesizeList []string
	for _, i := range strings.Split(hwPagesize, "\n") {
		if len(i) > 1 {
			hwPagesizeList = append(hwPagesizeList, i)
		}
	}
	hwPagesize = hwPagesizeList[0]

	var vFreeCountInt, _ = strconv.Atoi(vFreeCount)
	var hwPagesizeInt, _ = strconv.Atoi(hwPagesize)

	var finalResult = vFreeCountInt * hwPagesizeInt

	return strconv.Itoa(finalResult)
}

func ByteConversion(bytes int) string {
	return "Pass"
}
