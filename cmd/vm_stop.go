package cmd

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"

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

	stopBhyveProcess(vmName)
	vmSupervisorCleanup(vmName)
	networkCleanup(vmName)

	return nil
}

func stopBhyveProcess(vmName string) {
	prepCmd1 := "pgrep"
	prepCmd2 := "-lf"
	prepCmd3 := vmName
	cmd := exec.Command(prepCmd1, prepCmd2, prepCmd3)
	stdout, stderr := cmd.Output()
	if stderr != nil {
		if cmd.ProcessState.ExitCode() == 1 {
			_ = 0
		} else {
			log.Fatal("pgrep exited with an error " + stderr.Error())
		}
	}

	processId := ""
	reMatchVm, _ := regexp.Compile(`.*bhyve:\s` + vmName)
	for _, v := range strings.Split(string(stdout), "\n") {
		if reMatchVm.MatchString(v) {
			processId = strings.TrimSpace(strings.Split(v, " ")[0])
		}
	}
	stopCommand1 := "kill"
	stopCommand2 := "-SIGTERM"
	cmd = exec.Command(stopCommand1, stopCommand2, processId)
	stderr = cmd.Run()
	if stderr != nil {
		log.Fatal("kill was not successful " + stderr.Error())
	}
	fmt.Println("kill -SIGKILL " + processId)
}

func vmSupervisorCleanup(vmName string) {
	reMatchVm, _ := regexp.Compile(`for\s` + vmName + `\s&`)
	processId := ""

	prepCmd1 := "pgrep"
	prepCmd2 := "-lf"
	prepCmd3 := vmName

	iteration := 0
	for {
		time.Sleep(time.Second * 2)

		processId = ""
		cmd := exec.Command(prepCmd1, prepCmd2, prepCmd3)
		stdout, stderr := cmd.Output()
		if stderr != nil {
			if cmd.ProcessState.ExitCode() == 1 {
				_ = 0
			} else {
				log.Fatal("pgrep exited with an error " + stderr.Error())
			}
		}

		for _, v := range strings.Split(string(stdout), "\n") {
			v = strings.TrimSpace(v)
			if reMatchVm.MatchString(v) {
				processId = strings.Split(v, " ")[0]
			}
		}

		if len(processId) < 1 {
			fmt.Println("Process is gonzo")
			break
		}

		iteration = iteration + 1
		if iteration > 3 {
			stopCommand1 := "kill"
			stopCommand2 := "-SIGKILL"
			cmd := exec.Command(stopCommand1, stopCommand2, processId)
			stderr := cmd.Run()
			if stderr != nil {
				log.Fatal("kill was not successful " + stderr.Error())
			}
			fmt.Println("kill -SIGKILL " + processId)
			break
		}
	}
}

func networkCleanup(vmName string) {
	fmt.Println("Starting network cleanup")
	cmd := exec.Command("ifconfig")
	stdout, stderr := cmd.Output()
	if stderr != nil {
		log.Fatal("ifconfig exited with an error " + stderr.Error())
	}

	reMatchDescription, _ := regexp.Compile(`.*description:.*`)
	reMatchVm, _ := regexp.Compile(`\s+` + vmName + `\s+`)
	rePickTap, _ := regexp.Compile(`[\s|"]tap\d+`)
	for _, v := range strings.Split(string(stdout), "\n") {
		if reMatchDescription.MatchString(v) && reMatchVm.MatchString(v) {
			tap := rePickTap.FindString(v)
			tap = strings.TrimSpace(tap)
			tap = strings.ReplaceAll(tap, "\"", "")
			fmt.Println("ifconfig " + tap + " destroy")
		}
	}
}
