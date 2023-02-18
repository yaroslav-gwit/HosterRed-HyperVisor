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
	snapshotType    string
	snapshotsToKeep int

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

// Snapshot a given VM. Returns an error, if something wrong happened in the process.
func vmZfsSnapshot(vmName string) error {
	possibleSnapshotTypes := []string{"hourly", "daily", "weekly", "monthly", "yearly", "custom"}
	if !slices.Contains(possibleSnapshotTypes, snapshotType) {
		return errors.New("this snapshot type `" + snapshotType + "` is not supported by our system")
	}

	vmDataset, err := getVmDataset(vmName)
	if err != nil {
		return errors.New("getVmDataset(vmName): " + err.Error())
	}
	fmt.Println("Working with this VM dataset: " + vmDataset)

	vmSnapshotList, err := getVmSnapshots(vmDataset)
	if err != nil {
		return errors.New("getVmSnapshots(vmDataset) exited with an error: " + err.Error())
	}
	fmt.Println("VM snapshot list:")
	for _, v := range vmSnapshotList {
		fmt.Println(v)
	}

	err = takeNewSnapshot(vmDataset, snapshotType)
	if err != nil {
		return errors.New("takeNewSnapshot() exited with an error: " + err.Error())
	}
	return nil
}

// Runs `zfs list` command to return an active VM dataset.
// Useful for taking snapshots, cloning and destroying the VMs.
func getVmDataset(vmName string) (string, error) {
	zfsListCmd1 := "zfs"
	zfsListCmd2 := "list"
	zfsListCmd3 := "-H"

	cmd := exec.Command(zfsListCmd1, zfsListCmd2, zfsListCmd3)
	stdout, stderr := cmd.Output()
	if stderr != nil {
		return "", errors.New("zfs list exited with an error: " + stderr.Error())
	}

	reVmMatch := regexp.MustCompile(`.*/` + vmName + `\s`)
	reDsSplit := regexp.MustCompile(`\s+`)

	var result string
	for _, v := range strings.Split(string(stdout), "\n") {
		v = strings.TrimSpace(v)
		if reVmMatch.MatchString(v) {
			result = reDsSplit.Split(v, -1)[0]
			break
		}
	}

	if len(result) < 1 {
		return "", errors.New("can't find the dataset for this VM, sorry")
	}

	return result, nil
}

// Returns the current list of VM snapshots
func getVmSnapshots(vmDataset string) ([]string, error) {
	var listOfSnaps []string
	zfsListCmd1 := "zfs"
	zfsListCmd2 := "list"
	zfsListCmd3 := "-t"
	zfsListCmd4 := "snapshot"
	zfsListCmd5 := "-Hp"

	cmd := exec.Command(zfsListCmd1, zfsListCmd2, zfsListCmd3, zfsListCmd4, zfsListCmd5, vmDataset)
	stdout, stderr := cmd.Output()
	if stderr != nil {
		return listOfSnaps, errors.New("zfs list exited with an error: " + stderr.Error())
	}

	reDsSplit := regexp.MustCompile(`\s+`)
	for _, v := range strings.Split(string(stdout), "\n") {
		v = strings.TrimSpace(v)
		listOfSnaps = append(listOfSnaps, reDsSplit.Split(v, -1)[0])
	}

	return listOfSnaps, nil
}

func takeNewSnapshot(vmDataset string, snapshotType string) error {
	zfsSnapCmd1 := "zfs"
	zfsSnapCmd2 := "snapshot"

	now := time.Now()
	timeNow := now.Format("2006-01-02_15-04-05")
	cmd := exec.Command(zfsSnapCmd1, zfsSnapCmd2, vmDataset+"@"+snapshotType+"_"+timeNow)
	err := cmd.Run()
	if err != nil {
		return errors.New("zfs snapshot exited with an error: " + err.Error())
	}

	return nil
}
