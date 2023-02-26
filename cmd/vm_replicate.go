package cmd

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	replicationEndpoint string
	endpointSshPort     int
	sshKeyLocation      string

	vmZfsReplicateCmd = &cobra.Command{
		Use:   "replicate [vmName]",
		Short: "Use ZFS replication to send this VM to another host",
		Long:  `Use ZFS replication to send this VM to another host`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(replicationEndpoint) < 1 {
				log.Fatal("Please specify an endpoint!")
			}
			vmName := args[0]
			err := replicateVm(vmName, replicationEndpoint, endpointSshPort, sshKeyLocation)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func replicateVm(vmName string, replicationEndpoint string, endpointSshPort int, sshKeyLocation string) error {
	if !slices.Contains(getAllVms(), vmName) {
		return errors.New("vm does not exist")
	}

	_, err := checkSshConnection(replicationEndpoint, endpointSshPort, sshKeyLocation)
	if err != nil {
		return err
	}

	zfsDatasets, err := getRemoteZfsDatasets(replicationEndpoint, endpointSshPort, sshKeyLocation)
	if err != nil {
		return err
	}

	reMatchVm := regexp.MustCompile(`.*/` + vmName + `$`)
	reMatchVmSnaps := regexp.MustCompile(`.*/` + vmName + `@.*`)

	var remoteVmDataset []string
	var remoteVmSnapshots []string
	for _, v := range zfsDatasets {
		v = strings.TrimSpace(v)
		if reMatchVm.MatchString(v) {
			remoteVmDataset = append(remoteVmDataset, v)
		} else if reMatchVmSnaps.MatchString(v) {
			remoteVmSnapshots = append(remoteVmSnapshots, v)
		}
	}

	fmt.Println("Vm Remote Dataset:")
	fmt.Println(remoteVmDataset)
	fmt.Println()

	fmt.Println("Vm Local Snapshots:")
	vmDataset, err := getVmDataset(vmName)
	if err != nil {
		return err
	}
	localVmSnaps, err := getVmSnapshots(vmDataset)
	if err != nil {
		return err
	}
	fmt.Println(localVmSnaps)
	fmt.Println()

	var snapshotDiff []string
	fmt.Println("Vm Remote Snapshots:")
	fmt.Println(remoteVmSnapshots)
	fmt.Println()

	for _, v := range remoteVmSnapshots {
		if !slices.Contains(localVmSnaps, v) {
			snapshotDiff = append(snapshotDiff, v)
		}
	}
	fmt.Println("Vm Snapshot Diff:")
	fmt.Println(snapshotDiff)
	fmt.Println()

	return nil
}

const SshConnectionTimeout = "timeout"
const SshConnectionLoginFailure = "login failure"
const SshConnectionCantResolve = "cant resolve"
const SshConnectionSuccess = "success"

func checkSshConnection(replicationEndpoint string, endpointSshPort int, sshKeyLocation string) (string, error) {
	sshPort := strconv.Itoa(endpointSshPort)
	stdout, stderr := exec.Command("ssh", "-v", "-oConnectTimeout=2", "-oConnectionAttempts=2", "-oBatchMode=yes", "-i", sshKeyLocation, "-p"+sshPort, replicationEndpoint, "echo", "success").CombinedOutput()
	reMatchTimeout := regexp.MustCompile(`.*Operation timed out.*`)
	reMatchCantResolve := regexp.MustCompile(`.*Name does not resolve.*`)
	reMatchLoginFailure := regexp.MustCompile(`.*Permission denied.*`)
	if stderr != nil {
		if reMatchTimeout.MatchString(string(stdout)) {
			return "", errors.New("ssh connection error: " + SshConnectionTimeout)
		}
		if reMatchCantResolve.MatchString(string(stdout)) {
			return "", errors.New("ssh connection error: " + SshConnectionCantResolve)
		}
		if reMatchLoginFailure.MatchString(string(stdout)) {
			return "", errors.New("ssh connection error: " + SshConnectionLoginFailure)
		}
	}

	return SshConnectionSuccess, nil
}

func getRemoteZfsDatasets(replicationEndpoint string, endpointSshPort int, sshKeyLocation string) ([]string, error) {
	sshPort := strconv.Itoa(endpointSshPort)
	stdout, stderr := exec.Command("ssh", "-oBatchMode=yes", "-i", sshKeyLocation, "-p"+sshPort, replicationEndpoint, "zfs", "list", "-t", "all").CombinedOutput()
	if stderr != nil {
		return []string{}, errors.New("ssh connection error: " + string(stdout))
	}

	var remoteDatasetList []string
	reSplitSpace := regexp.MustCompile(`\s+`)
	for _, v := range strings.Split(string(stdout), "\n") {
		tempResult := reSplitSpace.Split(v, -1)[0]
		if len(tempResult) > 0 {
			remoteDatasetList = append(remoteDatasetList, tempResult)
		}
	}

	return remoteDatasetList, nil
}
