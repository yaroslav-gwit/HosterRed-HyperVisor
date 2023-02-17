package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	vmStartCmd = &cobra.Command{
		Use:   "start [vmName]",
		Short: "Start a particular VM using it's name",
		Long:  `Start a particular VM using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// cmd.Help()
			// fmt.Println(args[0])
			err := vmStart(args[0])
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func vmStart(vmName string) error {
	allVms := getAllVms()
	if slices.Contains(allVms, vmName) {
		bhyveCommand := generateBhyveStartCommand(vmName)
		vmProcessSupervisor(bhyveCommand)
	} else {
		return errors.New("VM is not found in the system")
	}

	return nil
}

func generateBhyveStartCommand(vmName string) string {
	vmConfigVar := vmConfig(vmName)

	var availableTaps []string
	for _, v := range vmConfigVar.Networks {
		availableTap := findAvailableTapInterface()
		availableTaps = append(availableTaps, availableTap)
		// fmt.Println("Next available tap int:", availableTap)

		createTapInterface := "ifconfig " + availableTap + " create"
		fmt.Println(createTapInterface)
		parts := strings.Fields(createTapInterface)
		exec.Command(parts[0], parts[1:]...).Run()

		bridgeTapInterface := "ifconfig vm-" + v.NetworkBridge + " addm " + availableTap
		fmt.Println(bridgeTapInterface)
		parts = strings.Fields(bridgeTapInterface)
		exec.Command(parts[0], parts[1:]...).Run()

		upBridgeInterface := "ifconfig vm-" + v.NetworkBridge + " up"
		fmt.Println(upBridgeInterface)
		parts = strings.Fields(upBridgeInterface)
		exec.Command(parts[0], parts[1:]...).Run()

		setTapDescription := "ifconfig " + availableTap + " description " + "\"" + availableTap + " " + vmName + " interface " + v.NetworkBridge + "\""
		fmt.Println(setTapDescription)
		parts = strings.Fields(setTapDescription)
		exec.Command(parts[0], parts[1:]...).Run()
	}

	bhyveFinalCommand := "bhyve -HAw -s 0:0,hostbridge -s 31,lpc "
	bhyvePci1 := 2
	bhyvePci2 := 0

	var networkFinal string
	var networkAdaptorType string
	if len(vmConfigVar.Networks) > 1 {
		for i, v := range vmConfigVar.Networks {
			networkAdaptorType = "," + v.NetworkAdaptorType + ","
			if i == 0 {
				networkFinal = "-s " + strconv.Itoa(bhyvePci1) + ":" + strconv.Itoa(bhyvePci2) + networkAdaptorType + availableTaps[i] + ",mac=" + v.NetworkMac
			} else {
				bhyvePci2 = bhyvePci2 + 1
				networkFinal = networkFinal + " -s " + strconv.Itoa(bhyvePci1) + ":" + strconv.Itoa(bhyvePci2) + networkAdaptorType + availableTaps[i] + ",mac=" + v.NetworkMac
			}
		}
	} else {
		networkAdaptorType = "," + vmConfigVar.Networks[0].NetworkAdaptorType + ","
		networkFinal = "-s " + strconv.Itoa(bhyvePci1) + ":" + strconv.Itoa(bhyvePci2) + networkAdaptorType + availableTaps[0] + ",mac=" + vmConfigVar.Networks[0].NetworkMac
	}

	bhyveFinalCommand = bhyveFinalCommand + networkFinal
	// fmt.Println(bhyveFinalCommand)

	bhyvePci := 3
	var diskFinal string
	var genericDiskText string
	var diskImageLocation string
	if len(vmConfigVar.Disks) > 1 {
		for i, v := range vmConfigVar.Disks {
			diskImageLocation = getVmFolder(vmName) + "/" + v.DiskImage
			genericDiskText = ":0," + v.DiskType + ","
			if i == 0 {
				diskFinal = " -s " + strconv.Itoa(bhyvePci) + genericDiskText + diskImageLocation
			} else {
				bhyvePci = bhyvePci + 1
				diskFinal = diskFinal + " -s " + strconv.Itoa(bhyvePci) + genericDiskText + diskImageLocation
			}
		}
	} else {
		diskImageLocation = getVmFolder(vmName) + "/" + vmConfigVar.Disks[0].DiskImage
		genericDiskText = ":0," + vmConfigVar.Disks[0].DiskType + ","
		diskFinal = " -s " + strconv.Itoa(bhyvePci) + genericDiskText + diskImageLocation
	}

	bhyveFinalCommand = bhyveFinalCommand + diskFinal
	// fmt.Println(bhyveFinalCommand)

	cpuAndRam := " -c sockets=" + vmConfigVar.CPUSockets + ",cores=" + vmConfigVar.CPUCores + " -m " + vmConfigVar.Memory
	bhyveFinalCommand = bhyveFinalCommand + cpuAndRam
	// fmt.Println(bhyveFinalCommand)

	bhyvePci = bhyvePci + 1
	vncCommand := " -s " + strconv.Itoa(bhyvePci) + ":" + strconv.Itoa(bhyvePci2) + ",fbuf,tcp=0.0.0.0:" + vmConfigVar.VncPort + ",w=1280,h=1024,password=" + vmConfigVar.VncPassword
	bhyveFinalCommand = bhyveFinalCommand + vncCommand
	// fmt.Println(bhyveFinalCommand)

	bhyvePci = bhyvePci + 1
	var loaderCommand string
	if vmConfigVar.Loader == "bios" {
		loaderCommand = " -s " + strconv.Itoa(bhyvePci) + ":" + strconv.Itoa(bhyvePci2) + ",xhci,tablet -l com1,/dev/nmdm-" + vmName + "-1A -l bootrom,/usr/local/share/uefi-firmware/BHYVE_UEFI_CSM.fd -u " + vmName
	} else if vmConfigVar.Loader == "uefi" {
		loaderCommand = " -s " + strconv.Itoa(bhyvePci) + ":" + strconv.Itoa(bhyvePci2) + ",xhci,tablet -l com1,/dev/nmdm-" + vmName + "-1A -l bootrom,/usr/local/share/uefi-firmware/BHYVE_UEFI.fd -u " + vmName
	} else {
		log.Fatal("Please make sure your loader is set to 'bios' or 'uefi'")
	}

	bhyveFinalCommand = bhyveFinalCommand + loaderCommand
	fmt.Println("Will execute this bhyve command: " + bhyveFinalCommand)

	return bhyveFinalCommand
}

func findAvailableTapInterface() string {
	cmd := exec.Command("ifconfig")
	stdout, stderr := cmd.Output()
	if stderr != nil {
		log.Fatal("ifconfig exited with an error " + stderr.Error())
	}

	reMatchTap, _ := regexp.Compile(`^tap`)

	var tapList []int
	var trimmedTap string
	for _, v := range strings.Split(string(stdout), "\n") {
		trimmedTap = strings.Trim(v, "")
		if reMatchTap.MatchString(trimmedTap) {
			for _, vv := range strings.Split(trimmedTap, ":") {
				if reMatchTap.MatchString(vv) {
					vv = strings.Replace(vv, "tap", "", 1)
					vvInt, err := strconv.Atoi(vv)
					if err != nil {
						log.Fatal("Could not convert tap int: " + err.Error())
					}
					tapList = append(tapList, vvInt)
				}
			}
		}
	}

	nextFreeTap := 0
	for {
		if slices.Contains(tapList, nextFreeTap) {
			nextFreeTap = nextFreeTap + 1
		} else {
			return "tap" + strconv.Itoa(nextFreeTap)
		}
	}
}

func vmProcessSupervisor(command string) {
	parts := strings.Fields(command)
	for {
		cmd := exec.Command(parts[0], parts[1:]...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatalf("Failed to create stdout pipe: %v", err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Fatalf("Failed to create stderr pipe: %v", err)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		stdoutReader := bufio.NewReader(stdout)
		go func() {
			defer wg.Done()
			readAndLogOutput(stdoutReader, "stdout")
		}()

		stderrReader := bufio.NewReader(stderr)
		go func() {
			defer wg.Done()
			readAndLogOutput(stderrReader, "stderr")
		}()

		done := make(chan error)
		startCommand(cmd, done)

		wg.Wait()

		if err := <-done; err != nil {
			log.Printf("Command failed: %v", err)
			if exitError, ok := err.(*exec.ExitError); ok {
				if status, ok := exitError.Sys().(interface{ ExitStatus() int }); ok {
					exitCode := status.ExitStatus()
					if exitCode != 100 {
						log.Printf("Command returned non-zero exit code: %d, restarting...", exitCode)
						continue
					}
				}
			}
			log.Fatal("Failed to get exit code")
		}

		time.Sleep(time.Second)
	}
}

func readAndLogOutput(reader *bufio.Reader, name string) {
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read %s: %v", name, err)
		}
		line = strings.TrimSpace(line)
		if line != "" {
			log.Printf("[%s] %s\n", name, line)
		}
	}
}

func startCommand(cmd *exec.Cmd, done chan error) {
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}
	go func() {
		done <- cmd.Wait()
	}()
}
