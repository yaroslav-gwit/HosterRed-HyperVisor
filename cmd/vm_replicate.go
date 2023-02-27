package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"hoster/emojlog"
	"log"
	"os"
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
	if len(remoteVmSnapshots) > 0 {
		emojlog.PrintLogMessage("Working with this remote dataset: "+remoteVmDataset[0], emojlog.Info)
	}

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

	if len(remoteVmSnapshots) < 1 {
		err = sendInitialSnapshot(vmDataset, localVmSnaps[0], replicationEndpoint, endpointSshPort, sshKeyLocation)
		if err != nil {
			return err
		}
	}

	if len(remoteVmSnapshots) > 0 {
		emojlog.PrintLogMessage("Replication for "+remoteVmDataset[0]+" is now finished", emojlog.Info)
	}

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

func sendInitialSnapshot(endpointDataset string, snapshotToSend string, replicationEndpoint string, endpointSshPort int, sshKeyLocation string) error {
	emojlog.PrintLogMessage("Starting replication for "+snapshotToSend, emojlog.Debug)

	out, err := exec.Command("zfs", "send", "-nP", snapshotToSend).CombinedOutput()
	if err != nil {
		return err
	}

	reMatchSize := regexp.MustCompile(`^size.*`)
	reMatchWhitespace := regexp.MustCompile(`\s+`)
	reMatchTime := regexp.MustCompile(`.*\d\d:\d\d:\d\d.*`)

	var snapshotSize int
	for _, v := range strings.Split(string(out), "\n") {
		if reMatchSize.MatchString(v) {
			tempInt, _ := strconv.Atoi(reMatchWhitespace.Split(v, -1)[1])
			snapshotSize = int(tempInt)
		}
	}

	bar := progressbar.NewOptions(
		snapshotSize,
		progressbar.OptionShowBytes(true),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetDescription(" 📤 Running ZFS send || "+snapshotToSend+" || "),
	)

	bashScript := []byte("zfs send -Pv " + snapshotToSend + " | ssh -i " + sshKeyLocation + " " + replicationEndpoint + " zfs receive -F " + endpointDataset)
	err = os.WriteFile("/tmp/replication.sh", bashScript, 0600)
	if err != nil {
		return err
	}

	cmd := exec.Command("sh", "/tmp/replication.sh")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	// read stderr output line by line and update the progress bar, parsing the line sting
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if reMatchTime.MatchString(line) {
			tempResult, _ := strconv.Atoi(reMatchWhitespace.Split(line, -1)[1])
			bar.Set(tempResult)
		}
	}

	// wait for command to finish
	if err := cmd.Wait(); err != nil {
		return err
	}

	bar.Finish()
	fmt.Println()
	emojlog.PrintLogMessage("Replication done for "+snapshotToSend, emojlog.Info)

	return nil
}
