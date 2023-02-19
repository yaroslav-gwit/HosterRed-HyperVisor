package cmd

import (
	"os"
	"os/exec"

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
	// tailCmd := exec.Command("micro", vmFolder+"/vm_supervisor.log")
	// cmd := exec.Command("bash", "-i")
	tailCmd := exec.Command("tail", "-n", "35", "-f", vmFolder+"/vm_supervisor.log")
	tailCmd.Stdin = os.Stdin
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr
	tailCmd.Run()

	// Start the command with a new pseudo-terminal
	// ptmx, err := pty.Start(tailCmd)
	// if err != nil {
	// 	panic(err)
	// }

	// Forward input/output to the terminal
	// go func() {
	// 	_, _ = io.Copy(ptmx, os.Stdin)
	// }()
	// go func() {
	// 	_, _ = io.Copy(os.Stdout, ptmx)
	// }()

	// Wait for the command to exit
	// if err := tailCmd.Wait(); err != nil {
	// 	_ = err
	// fmt.Println("Command failed:", err)
	// }
}
