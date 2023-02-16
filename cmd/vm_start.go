package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	vmStartCmd = &cobra.Command{
		Use:   "start [VM name]",
		Short: "Start a particular VM using it's name",
		Long:  `Start a particular VM using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// cmd.Help()
			// fmt.Println(args[0])
			vmStart(args[0])
		},
	}
)

func vmStart(vmName string) error {
	allVms := getAllVms()
	generateBhyveStartCommand(vmName)
	if slices.Contains(allVms, vmName) {
		_ = ""
	} else {
		return errors.New("VM is not found in the system")
	}

	return nil
}

func generateBhyveStartCommand(vmName string) string {
	// Find existing TAP adaptors
	cmd := exec.Command("ifconfig")
	stdout, stderr := cmd.Output()
	if stderr != nil {
		log.Fatal("ifconfig exited with an error " + stderr.Error())
	}

	reMatchTap, _ := regexp.Compile(`^tap`)

	var trimmedTap string
	// var tapList []string
	// nextFreeTap := 0
	for _, v := range strings.Split(string(stdout), "\n") {
		trimmedTap = strings.Trim(v, "")
		if reMatchTap.MatchString(trimmedTap) {
			for _, vv := range strings.Split(trimmedTap, ":") {
				if reMatchTap.MatchString(vv) {
					vv = strings.Replace(vv, "tap", "", 1)
					fmt.Println(vv)
				}
			}
		}
	}

	return ""
}

func test() {
	for {
		cmd := exec.Command("your-command")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		// Create pipes for stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Printf("Error creating stdout pipe: %s\n", err)
			os.Exit(1)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			fmt.Printf("Error creating stderr pipe: %s\n", err)
			os.Exit(1)
		}

		// Start the child process
		err = cmd.Start()
		if err != nil {
			fmt.Printf("Error starting command: %s\n", err)
			os.Exit(1)
		}

		// Read from stdout and stderr
		stdoutScanner := bufio.NewScanner(stdout)
		stderrScanner := bufio.NewScanner(stderr)
		go func() {
			for stdoutScanner.Scan() {
				fmt.Println(stdoutScanner.Text())
			}
		}()
		go func() {
			for stderrScanner.Scan() {
				fmt.Println(stderrScanner.Text())
			}
		}()

		// Wait for the child process to exit
		err = cmd.Wait()
		if err != nil {
			fmt.Printf("Error waiting for command: %s\n", err)
			os.Exit(1)
		} else {
			fmt.Println("Process exited")
		}
	}
}
