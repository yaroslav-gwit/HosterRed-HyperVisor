package cmd

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	vmStopCmd = &cobra.Command{
		Use:   "stop [vmName]",
		Short: "Stop a particular VM using it's name",
		Long:  `Stop a particular VM using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// cmd.Help()
			// fmt.Println(args[0])
			err := vmStop(args[0])
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func vmStop(vmName string) error {
	allVms := getAllVms()
	if !slices.Contains(allVms, vmName) {
		return errors.New("VM is not found on this system")
	} else if !vmLiveCheck(vmName) {
		return errors.New("VM is already stopped")
	}

	// stopCommand1 := "kill"
	// stopCommand2 := "-SIGTERM"
	// stopCommand3 := "74972"

	prepCmd1 := "pgrep"
	prepCmd2 := "-lf"
	prepCmd3 := vmName
	cmd := exec.Command(prepCmd1, prepCmd2, prepCmd3)
	stdout, stderr := cmd.Output()
	if stderr != nil {
		if cmd.ProcessState.ExitCode() == 1 {
			_ = 0
		} else {
			log.Fatal("ifconfig exited with an error " + stderr.Error())
		}
	}
	reMatchVm, _ := regexp.Compile(`.*bhyve:.*`)

	var processId string
	for _, v := range strings.Split(string(stdout), "\n") {
		if len(v) > 0 {
			if reMatchVm.MatchString(v) {
				processId = strings.Split(v, " ")[0]
			}
			fmt.Println(processId)
		}
	}

	return nil
}

// func findVmTapInterfaces() []string {
// 	cmd := exec.Command("ifconfig")
// 	stdout, stderr := cmd.Output()
// 	if stderr != nil {
// 		log.Fatal("ifconfig exited with an error " + stderr.Error())
// 	}

// 	reMatchTap, _ := regexp.Compile(`^tap`)

// 	var tapList []int
// 	var trimmedTap string
// 	for _, v := range strings.Split(string(stdout), "\n") {
// 		trimmedTap = strings.Trim(v, "")
// 		if reMatchTap.MatchString(trimmedTap) {
// 			for _, vv := range strings.Split(trimmedTap, ":") {
// 				if reMatchTap.MatchString(vv) {
// 					vv = strings.Replace(vv, "tap", "", 1)
// 					vvInt, err := strconv.Atoi(vv)
// 					if err != nil {
// 						log.Fatal("Could not convert tap int: " + err.Error())
// 					}
// 					tapList = append(tapList, vvInt)
// 				}
// 			}
// 		}
// 	}

// 	nextFreeTap := 0
// 	for {
// 		if slices.Contains(tapList, nextFreeTap) {
// 			nextFreeTap = nextFreeTap + 1
// 		} else {
// 			return "tap" + strconv.Itoa(nextFreeTap)
// 		}
// 	}
// 	return
// }
