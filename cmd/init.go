package cmd

import (
	"errors"
	"hoster/emojlog"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize all FreeBSD kernel modules required by hoster",
		Long:  `Initialize all FreeBSD kernel modules required by hoster`,
		Run: func(cmd *cobra.Command, args []string) {
			err := loadMissingModules()
			if err != nil {
				log.Fatal(err.Error())
			}
			err = applySysctls()
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

// kldload vmm
// kldload nmdm
// kldload if_bridge
// kldload if_tuntap
// kldload if_tap
// kldload pf
// kldload pflog
// sysctl net.link.tap.up_on_open=1

func loadMissingModules() error {
	moduleList, err := returnMissingModules()
	if err != nil {
		return err
	}

	for _, v := range moduleList {
		stdout, stderr := exec.Command("kldload", v).CombinedOutput()
		if stderr != nil {
			return errors.New("error running kldstat: " + string(stdout) + " " + stderr.Error())
		}
		emojlog.PrintLogMessage("Loaded kernel module: "+v, emojlog.Changed)
	}

	return nil
}

// Returns a list of kernel modules that are not yet loaded, or an error
func returnMissingModules() ([]string, error) {
	stdout, stderr := exec.Command("kldstat", "-v").CombinedOutput()
	if stderr != nil {
		return []string{}, errors.New("error running kldstat: " + string(stdout) + " " + stderr.Error())
	}

	reMatchKo := regexp.MustCompile(`\.ko`)
	reSplitSpace := regexp.MustCompile(`\s+`)

	var loadedModules []string
	kernelModuleList := []string{"vmm", "nmdm", "if_bridge", "pf", "pflog"}
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
							loadedModules = append(loadedModules, strings.TrimSpace(vvv))
						}
					}
				}
			}
		}
	}

	kernelModuleListNoKo := []string{"if_tuntap", "if_tap"}
	for _, v := range strings.Split(string(stdout), "\n") {
		for _, vv := range kernelModuleListNoKo {
			reMatchModule := regexp.MustCompile(vv)
			if reMatchModule.MatchString(v) {
				tempList := reSplitSpace.Split(v, -1)
				for _, vvv := range tempList {
					if reMatchModule.MatchString(vvv) {
						loadedModules = append(loadedModules, strings.TrimSpace(vvv))
					}
				}
			}
		}
	}

	stdout, stderr = exec.Command("sysctl", "net.link.tap.up_on_open").CombinedOutput()
	if stderr != nil {
		return []string{}, errors.New("error running sysctl: " + string(stdout) + " " + stderr.Error())
	}
	for _, v := range loadedModules {
		emojlog.PrintLogMessage("Module is already active: "+v, emojlog.Debug)
	}

	var result []string
	for _, v := range kernelModuleList {
		if !slices.Contains(loadedModules, v) {
			result = append(result, v)
		}
	}

	for _, v := range kernelModuleListNoKo {
		if !slices.Contains(loadedModules, v) {
			result = append(result, v)
		}
	}

	return result, nil
}

func applySysctls() error {
	stdout, stderr := exec.Command("sysctl", "net.link.tap.up_on_open").CombinedOutput()
	if stderr != nil {
		return errors.New("error running sysctl: " + string(stdout) + " " + stderr.Error())
	}

	reSplitSpace := regexp.MustCompile(`\s+`)

	tapUpOnOpen := false
	for _, v := range reSplitSpace.Split(string(stdout), -1) {
		if v == "1" {
			tapUpOnOpen = true
			emojlog.PrintLogMessage("Sysctl net.link.tap.up_on_open is already active", emojlog.Debug)
		}
	}

	if !tapUpOnOpen {
		err := exec.Command("sysctl", "net.link.tap.up_on_open=1").Run()
		if err != nil {
			return errors.New("error running sysctl: " + string(stdout) + " " + stderr.Error())
		}
		emojlog.PrintLogMessage("Applied: sysctl net.link.tap.up_on_open=1", emojlog.Changed)

	}

	return nil
}
