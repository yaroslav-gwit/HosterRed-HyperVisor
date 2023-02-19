package cmd

import (
	"io"
	"os"
	"os/exec"

	"github.com/kr/pty"
	"github.com/spf13/cobra"
)

var (
	vmShowLogCmd = &cobra.Command{
		Use:   "show-log [vmName]",
		Short: "Show log in real time using `tail -f`",
		Long:  `Show log in real time using "tail -f"`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			viewLog(args[0])
		},
	}
)

func viewLog(vmName string) {
	vmFolder := getVmFolder(vmName)
	tailCmd := exec.Command("tail", "-n", "35", "-f", vmFolder+"/vm_supervisor.log")

	// Start the command with a new pseudo-terminal
	ptmx, err := pty.Start(tailCmd)
	if err != nil {
		panic(err)
	}

	// Forward input/output to the terminal
	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
	}()
	go func() {
		_, _ = io.Copy(os.Stdout, ptmx)
	}()

	// Wait for the command to exit
	if err := tailCmd.Wait(); err != nil {
		_ = err
		// fmt.Println("Command failed:", err)
	}
}
