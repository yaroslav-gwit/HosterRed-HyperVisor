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

	StopBhyveProcess(vmName)
	vmSupervisorCleanup(vmName)
	StopBhyveProcess(vmName)
	NetworkCleanup(vmName)
	BhyvectlDestroy(vmName)

	return nil
}

func StopBhyveProcess(vmName string) {
	fmt.Println("Stopping the VM")
	prepCmd1 := "pgrep"
	prepCmd2 := "-lf"
	prepCmd3 := vmName
	cmd := exec.Command(prepCmd1, prepCmd2, prepCmd3)
	stdout, stderr := cmd.Output()
	if stderr != nil {
		if cmd.ProcessState.ExitCode() == 1 {
			_ = 0
		} else {
			log.Println("pgrep exited with an error " + stderr.Error())
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
		log.Println("kill was not successful " + stderr.Error())
	}

	fmt.Println("Done stopping the VM")
}

func vmSupervisorCleanup(vmName string) {
	fmt.Println("Starting vm supervisor cleanup")
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
				log.Println("pgrep exited with an error " + stderr.Error())
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
				log.Println("kill was not successful " + stderr.Error())
			}
			fmt.Println("Forcefully killing the vm_supervisor, due to operation timeout " + processId)
		}
	}
	fmt.Println("Done cleaning up after vm supervisor")
}

func NetworkCleanup(vmName string) {
	fmt.Println("Starting network cleanup")
	cmd := exec.Command("ifconfig")
	stdout, stderr := cmd.Output()
	if stderr != nil {
		log.Println("ifconfig exited with an error " + stderr.Error())
	}

	reMatchDescription, _ := regexp.Compile(`.*description:.*`)
	reMatchVm, _ := regexp.Compile(`\s+` + vmName + `\s+`)
	rePickTap, _ := regexp.Compile(`[\s|"]tap\d+`)
	for _, v := range strings.Split(string(stdout), "\n") {
		if reMatchDescription.MatchString(v) && reMatchVm.MatchString(v) {
			tap := rePickTap.FindString(v)
			tap = strings.TrimSpace(tap)
			tap = strings.ReplaceAll(tap, "\"", "")
			fmt.Println("Destroying " + tap)
			ifconfigDestroyCmd1 := "ifconfig"
			ifconfigDestroyCmd3 := "destroy"
			cmd := exec.Command(ifconfigDestroyCmd1, tap, ifconfigDestroyCmd3)
			stderr := cmd.Run()
			if stderr != nil {
				log.Println("ifconfig destroy was not successful " + stderr.Error())
			}
		}
	}
	fmt.Println("Done cleaning up TAP network interfaces")
}

func BhyvectlDestroy(vmName string) {
	fmt.Println("Cleaning up Bhyve resources")
	time.Sleep(time.Second * 2)
	lsCommand1 := "ls"
	lsCommand2 := "-1"
	lsCommand3 := "/dev/vmm/"
	cmd := exec.Command(lsCommand1, lsCommand2, lsCommand3)
	stdout, _ := cmd.Output()

	matchVM, _ := regexp.Compile(`^` + vmName + `$`)
	for _, v := range strings.Split(string(stdout), "\n") {
		v = strings.TrimSpace(v)
		if matchVM.MatchString(v) {
			fmt.Println("Destroying a VM using bhyvectl: " + vmName)
			bhyvectlCommand1 := "bhyvectl"
			bhyvectlCommand2 := "--destroy"
			bhyvectlCommand3 := "--vm=" + vmName
			cmd := exec.Command(bhyvectlCommand1, bhyvectlCommand2, bhyvectlCommand3)
			stderr := cmd.Run()
			if stderr != nil {
				log.Println("bhyvectl exited with an error " + stderr.Error())
			}
		}
	}
	fmt.Println("Done cleaning up Bhyve resources")
}
