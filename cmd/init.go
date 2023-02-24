package cmd

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize all FreeBSD kernel modules required by hoster",
		Long:  `Initialize all FreeBSD kernel modules required by hoster`,
		Run: func(cmd *cobra.Command, args []string) {
			result, err := returnMissingModules()
			if err != nil {
				log.Fatal(err.Error())
			}
			for _, v := range result {
				fmt.Println(v)
			}
		},
	}
)

// #_ LIST OF MODULES TO LOAD _#
// kldstat -m $MODULE
// kldstat -mq $MODULE
// kldload vmm
// kldload nmdm
// kldload if_bridge
// kldload if_tuntap
// kldload if_tap
// sysctl net.link.tap.up_on_open=1
// 13.0-RELEASE-p11

// Returns a list modules that are not yet loaded, or an error
func returnMissingModules() ([]string, error) {
	var result []string
	stdout, stderr := exec.Command("kldstat", "-v").Output()
	if stderr != nil {
		return []string{}, stderr
	}

	for _, v := range strings.Split(string(stdout), "\n") {
		result = append(result, strings.TrimSpace(v))
	}

	return result, nil
}
