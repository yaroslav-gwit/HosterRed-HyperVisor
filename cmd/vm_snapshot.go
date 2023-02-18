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
	vmZfsSnapshotCmd = &cobra.Command{
		Use:   "snapshot",
		Short: "Snapshot running or offline VM",
		Long:  `Snapshot running or offline VM`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := vmZfsSnapshot(args[0])
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func vmZfsSnapshot(vmName string) error {
	fmt.Println("ZSF Snapshot")
	fmt.Println(getVmDataset(vmName))
	return nil
}

func getVmDataset(vmName string) (string, error) {
	zfsListCmd1 := "zfs"
	zfsListCmd2 := "list"
	zfsListCmd3 := "-H"

	cmd := exec.Command(zfsListCmd1, zfsListCmd2, zfsListCmd3)
	stdout, stderr := cmd.Output()
	if stderr != nil {
		if cmd.ProcessState.ExitCode() == 1 {
			_ = 0
		} else {
			log.Println("zfs list exited with an error " + stderr.Error())
		}
	}

	reVmMatch := regexp.MustCompile(`.*/` + vmName + `\s`)

	var result string
	for _, v := range strings.Split(string(stdout), "\n") {
		v = strings.TrimSpace(v)
		if reVmMatch.MatchString(v) {
			result = strings.Split(v, " ")[0]
		}
	}

	return result, nil
}
