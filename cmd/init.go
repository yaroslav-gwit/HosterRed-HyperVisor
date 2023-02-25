package cmd

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
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
// kldload pf
// kldload pflog
// sysctl net.link.tap.up_on_open=1
// 13.0-RELEASE-p11

// Returns a list modules that are not yet loaded, or an error
func returnMissingModules() ([]string, error) {
	var result []string
	stdout, stderr := exec.Command("kldstat", "-v").Output()
	if stderr != nil {
		return []string{}, stderr
	}

	reMatchKo := regexp.MustCompile(`\.ko`)
	reSplitSpace := regexp.MustCompile(`\s+`)
	kernelModuleList := []string{"vmm", "nmdm", "if_bridge", "pf", "pflog"}
	kernelModuleListNoKo := []string{"if_tuntap", "if_tap"}

	for _, v := range strings.Split(string(stdout), "\n") {
		if reMatchKo.MatchString(v) {
			for _, vv := range kernelModuleList {
				reMatchModule := regexp.MustCompile(vv + `\.ko`)
				reMatchModuleFinal := regexp.MustCompile(vv + `\.ko$`)
				if reMatchModule.MatchString(v) {
					tempList := reSplitSpace.Split(v, -1)
					for _, vvv := range tempList {
						if reMatchModuleFinal.MatchString(vvv) {
							vvv = reMatchKo.ReplaceAllString(vvv, "")
							result = append(result, strings.TrimSpace(vvv))
						}
					}
				}
			}
		}
	}

	for _, v := range strings.Split(string(stdout), "\n") {
		for _, vv := range kernelModuleListNoKo {
			reMatchModule := regexp.MustCompile(vv)
			if reMatchModule.MatchString(v) {
				tempList := reSplitSpace.Split(v, -1)
				for _, vvv := range tempList {
					if reMatchModule.MatchString(vvv) {
						result = append(result, strings.TrimSpace(vvv))
					}
				}
			}
		}
	}

	return result, nil
}
