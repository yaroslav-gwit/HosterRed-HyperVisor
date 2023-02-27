package cmd

import (
	"errors"
	"fmt"
	"hoster/emojlog"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/schollz/progressbar/v3"
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
	emojlog.PrintLogMessage("Working with this remote dataset: "+remoteVmDataset[0], emojlog.Info)

	vmDataset, err := getVmDataset(vmName)
	if err != nil {
		return err
	}
	localVmSnaps, err := getVmSnapshots(vmDataset)
	if err != nil {
		return err
	}
	var snapshotDiff []string
	for _, v := range remoteVmSnapshots {
		if !slices.Contains(localVmSnaps, v) {
			snapshotDiff = append(snapshotDiff, v)
		}
	}
	if len(snapshotDiff) > 0 {
		snapshotDiffStr := fmt.Sprint("Will be removing these snapshots: ", snapshotDiff)
		emojlog.PrintLogMessage(snapshotDiffStr, emojlog.Info)
		for _, v := range snapshotDiff {
			sshPort := strconv.Itoa(endpointSshPort)
			stdout, stderr := exec.Command("ssh", "-oBatchMode=yes", "-i", sshKeyLocation, "-p"+sshPort, replicationEndpoint, "zfs", "destroy", v).CombinedOutput()
			if stderr != nil {
				return errors.New("ssh connection error: " + string(stdout))
			}
			emojlog.PrintLogMessage("Destroyed an old snapshot: "+v, emojlog.Changed)
		}
	}

	sendSnapshot()

	emojlog.PrintLogMessage("Replication for "+remoteVmDataset[0]+" is now finished", emojlog.Info)
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

func sendSnapshot() {
	// Set the local dataset to replicate
	localDataset := "zroot/vm-encrypted/replicationTestVm"

	// Set the SSH command to run on the remote system
	remoteDataset := "zroot/vm-encrypted/replicationTestVm"

	// Set the SSH options
	sshHost := "192.168.120.17"
	// sshUser := "username"
	sshKey := "/root/.ssh/id_rsa"

	// Build the SSH command string
	sshCmd := exec.Command("ssh", "-i", sshKey, sshHost, "zfs", "receive", "-F", remoteDataset)

	// Build the local zfs send command
	zfsCmd := exec.Command("zfs", "send", "-v", "-p", localDataset)

	// Set up a progress bar for the zfsCmd
	zfsStats, err := zfsCmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	bar := progressbar.DefaultBytes(
		-1, // Set the total size to unknown
		"Replicating "+localDataset+": ",
	)

	// Set the Stdout of the zfsCmd to the Stdin of the sshCmd
	zfsOut, err := zfsCmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	sshCmd.Stdin = zfsOut

	// Start the zfsCmd and sshCmd
	if err := zfsCmd.Start(); err != nil {
		panic(err)
	}
	if err := sshCmd.Start(); err != nil {
		panic(err)
	}

	// Read output from zfsCmd and update the progress bar
	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := zfsStats.Read(buf)
			if err != nil {
				break
			}
			bar.Add(n)
		}
	}()

	// Wait for the commands to finish
	if err := zfsCmd.Wait(); err != nil {
		panic(err)
	}
	if err := sshCmd.Wait(); err != nil {
		panic(err)
	}

	fmt.Println("Replication completed successfully!")
}
