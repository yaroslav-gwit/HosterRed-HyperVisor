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
	var vFreeCount = "sysctl -nq vm.stats.vm.v_free_count"
	var cmd = exec.Command("bash", "-c", vFreeCount)
	var stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("There has been an error:", err)
		os.Exit(1)
	} else {
		vFreeCount = string(stdout)
	}

	var hwPagesize = "sysctl -nq hw.pagesize"
	cmd = exec.Command("bash", "-c", vFreeCount)
	stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("There has been an error:", err)
		os.Exit(1)
	} else {
		hwPagesize = string(stdout)
	}

	var vFreeCountInt, _ = strconv.Atoi(vFreeCount)
	var hwPagesizeInt, _ = strconv.Atoi(hwPagesize)

	var finalResult = vFreeCountInt * hwPagesizeInt

	return strconv.Itoa(finalResult)
}
