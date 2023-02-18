package cmd

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	vmZfsSnapshotCmd = &cobra.Command{
		Use:   "snapshot [vmName]",
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
	vmDataset, err := getVmDataset(vmName)
	fmt.Println(vmDataset)
	if err != nil {
		log.Println("zfs list exited with an error " + err.Error())
	}
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
			return "", errors.New("zfs list exited with an error " + stderr.Error())
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

	if len(result) < 1 {
		return "", errors.New("can't find the dataset for this VM, sorry")
	}

	return result, nil
}
