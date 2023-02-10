package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"facette.io/natsort"
	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

var (
	jsonOutputVm bool

	vmCmd = &cobra.Command{
		Use:   "vm",
		Short: "VM related operations",
		Long:  `VM related operations, ie VM deloyment, stopping/starting the VMs, etc`,
		Run: func(cmd *cobra.Command, args []string) {
			VmMain()
		},
	}
)

var allVmsUptime []string

func VmMain() {
	var wg = &sync.WaitGroup{}
	wg.Add(2)
	var vmInfo []string
	var thisHostName string
	go func() { defer wg.Done(); vmInfo = getAllVms() }()
	go func() { defer wg.Done(); thisHostName = GetHostName() }()
	wg.Wait()

	var ID = 0
	var vmLive string
	var vmEncrypted string
	var vmProduction string
	var vmConfigVar VmConfigStruct
	var cpuFinal string

	var t = table.New(os.Stdout)
	t.SetLineStyle(table.StyleBrightCyan)
	t.SetDividers(table.UnicodeRoundedDividers)
	t.SetAlignment(table.AlignCenter, //ID
		table.AlignLeft,   // VM Name
		table.AlignCenter, // VM Status
		table.AlignCenter, // CPU and RAM
		table.AlignCenter, // Main IP
		table.AlignCenter, // VNC Port
		table.AlignCenter, // VNC Password
		table.AlignCenter, // OS Comment
		table.AlignCenter, // VM Uptime
		table.AlignCenter, // OS Disk Used
		table.AlignCenter) // Description

	t.SetHeaders("ID",
		"VM Name",
		"VM Status",
		"CPU/RAM",
		"Main IP",
		"VNC Port",
		"VNC Password",
		"OS Comment",
		"VM Uptime",
		"OS Disk (Used)",
		"Description")

	t.SetHeaderStyle(table.StyleBold)

	for _, vmName := range vmInfo {
		getVmUptimeNew(vmName)
		wg.Add(2)
		var vmOsDiskFullSize string
		var vmOsDiskFree string
		go func() { defer wg.Done(); vmOsDiskFullSize = getOsDiskFullSize(vmName) }()
		go func() { defer wg.Done(); vmOsDiskFree = getOsDiskFree(vmName) }()
		wg.Wait()
		// var vmOsDiskFullSize = getOsDiskFullSize(vmName)
		// var vmOsDiskFree = getOsDiskFree(vmName)

		// getVmUptime(vmName)

		vmConfigVar = vmConfig(vmName)
		ID = ID + 1
		vmLive = vmLiveCheckString(vmName)
		if vmConfigVar.ParentHost != thisHostName {
			vmLive = "💾"
			vmConfigVar.Description = "💾⏩ " + vmConfigVar.ParentHost
		}
		if vmConfigVar.LiveStatus == "production" {
			vmProduction = "🔁"
		} else {
			vmProduction = ""
		}

		vmEncrypted = encryptionCheckString(vmName)
		var cpuCoresInt, _ = strconv.Atoi(vmConfigVar.CPUCores)
		var cpuSocketsInt, _ = strconv.Atoi(vmConfigVar.CPUSockets)
		cpuFinal = strconv.Itoa(cpuCoresInt * cpuSocketsInt)
		var vmUptimeVar = getVmUptimeNew(vmName)

		t.AddRow(strconv.Itoa(ID),
			vmName,
			vmLive+vmEncrypted+vmProduction,
			cpuFinal+"/"+vmConfigVar.Memory,
			vmConfigVar.Networks[0].IPAddress,
			vmConfigVar.VncPort,
			vmConfigVar.VncPassword,
			vmConfigVar.OsComment,
			vmUptimeVar,
			vmOsDiskFree+"/"+vmOsDiskFullSize,
			vmConfigVar.Description)
	}

	t.Render()
}

// type vmInfoStruct struct {
// 	vmName      string
// 	vmDataset   string
// 	vmFolder    string
// 	vmEncrypted string
// }

func getAllVms() []string {
	var zfsDatasets []string
	var configFileName = "/vm_config.json"
	zfsDatasets = append(zfsDatasets, "zroot/vm-encrypted")
	zfsDatasets = append(zfsDatasets, "zroot/vm-unencrypted")

	var vmListSorted []string

	for _, dataset := range zfsDatasets {
		var dsFolder = "/" + dataset + "/"
		var _, err = os.Stat(dsFolder)
		if !os.IsNotExist(err) {
			vmFolders, err := ioutil.ReadDir(dsFolder)

			if err != nil {
				fmt.Println("Error!", err)
				os.Exit(1)
			}

			for _, vmFolder := range vmFolders {
				info, _ := os.Stat(dsFolder + vmFolder.Name())
				if info.IsDir() {
					var _, err = os.Stat(dsFolder + vmFolder.Name() + configFileName)
					if !os.IsNotExist(err) {
						vmListSorted = append(vmListSorted, vmFolder.Name())
					}
				}
			}
		}
	}

	natsort.Sort(vmListSorted)
	return vmListSorted
}

func encryptionCheck(vmName string) bool {
	var zfsDatasets []string
	var dsFolder string
	var finalResponse bool
	zfsDatasets = append(zfsDatasets, "zroot/vm-encrypted")
	// zfsDatasets = append(zfsDatasets, "zroot/vm-unencrypted")

	for _, dataset := range zfsDatasets {
		dsFolder = "/" + dataset + "/"
		var _, err = os.Stat(dsFolder + vmName)
		if !os.IsNotExist(err) {
			finalResponse = true
			if finalResponse {
				break
			}
		} else {
			finalResponse = false
		}
	}
	return finalResponse
}

func encryptionCheckString(vmName string) string {
	var result = encryptionCheck(vmName)
	if result {
		return "🔒"
	} else {
		return ""
	}
}

func vmLiveCheck(vmName string) bool {
	var _, err = os.Stat("/dev/vmm/" + vmName)
	if !os.IsNotExist(err) {
		return true
	} else {
		return false
	}
}

func vmLiveCheckString(vmName string) string {
	var vmLive = vmLiveCheck(vmName)
	if vmLive {
		return "🟢"
	} else {
		return "🔴"
	}
}

func vmConfig(vmName string) VmConfigStruct {
	var configFile = getVmFolder(vmName) + "/vm_config.json"
	var jsonData = VmConfigStruct{}
	var content, err = ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("vmConfig Function Error: ", err)
	}
	if err := json.Unmarshal(content, &jsonData); err != nil {
		panic(err)
	}
	return jsonData
}

func getVmFolder(vmName string) string {
	var zfsDatasets []string
	var dsFolder string
	var finalResponse string
	zfsDatasets = append(zfsDatasets, "zroot/vm-encrypted")
	zfsDatasets = append(zfsDatasets, "zroot/vm-unencrypted")

	for _, dataset := range zfsDatasets {
		dsFolder = "/" + dataset + "/"
		var _, err = os.Stat(dsFolder + vmName)
		if !os.IsNotExist(err) {
			finalResponse = dsFolder + vmName
		}
	}
	return finalResponse
}

func getVmUptimeNew(vmName string) string {
	var vmsUptime []string
	if len(allVmsUptime) > 0 {
		vmsUptime = allVmsUptime
	} else {
		// println("allVmsUptime is empty!")
		var psEtime1 = "ps"
		var psEtime2 = "axwww"
		var psEtime3 = "-o"
		var psEtime4 = "etimes,command"

		var cmd = exec.Command(psEtime1, psEtime2, psEtime3, psEtime4)
		stdout, err := cmd.Output()
		if err != nil {
			log.Fatal("getVmUptimeNew Error: ", err)
		}
		allVmsUptime = strings.Split(string(stdout), "\n")
		vmsUptime = allVmsUptime
	}
	rexMatchVmName, _ := regexp.Compile(`.*bhyve: ` + vmName + `.*`)
	var finalResult string
	for i, v := range vmsUptime {
		if rexMatchVmName.MatchString(v) {
			v = strings.TrimSpace(v)
			v = strings.Split(v, " ")[0]

			var vmUptimeInt, _ = strconv.ParseInt(v, 10, 64)
			var secondsModulus = int(vmUptimeInt) % 60.0

			var minutesSince = (float64(vmUptimeInt) - float64(secondsModulus)) / 60.0
			var minutesModulus = int(minutesSince) % 60.0

			var hoursSince = (minutesSince - float64(minutesModulus)) / 60
			var hoursModulus = int(hoursSince) % 24

			var daysSince = (int(hoursSince) - hoursModulus) / 24

			finalResult = strconv.Itoa(daysSince) + "d "
			finalResult = finalResult + strconv.Itoa(hoursModulus) + "h "
			finalResult = finalResult + strconv.Itoa(minutesModulus) + "m "
			finalResult = finalResult + strconv.Itoa(secondsModulus) + "s"

			// fmt.Println(vmName, finalResult)
			break
		} else if i == (len(vmsUptime) - 1) {
			finalResult = "0s"
		}
	}
	return finalResult
}

func getVmUptime(vmName string) string {
	var pidFile = "/var/run/" + vmName + ".pid"
	var pidResult string
	var finalResult string
	var content, err = ioutil.ReadFile(pidFile)
	if err != nil {
		finalResult = "-"
	} else {
		pidResult = string(content)
		pidResult = strings.Replace(pidResult, "\n", "", -1)
	}

	if len(pidResult) > 0 {
		var vmUptime string
		var vmUptimeArg1 = "ps"
		var vmUptimeArg2 = "-o"
		var vmUptimeArg3 = "etimes="
		var vmUptimeArg4 = "-p"
		var vmUptimeArg5 = pidResult

		var cmd = exec.Command(vmUptimeArg1, vmUptimeArg2, vmUptimeArg3, vmUptimeArg4, vmUptimeArg5)
		stdout, err := cmd.Output()
		if err != nil {
			fmt.Println("Func getVmUptime: There has been an error:", err)
			// os.Exit(1)
		} else {
			vmUptime = string(stdout)
			vmUptime = strings.Replace(vmUptime, "\n", "", -1)
		}

		var vmUptimeInt, _ = strconv.ParseInt(vmUptime, 10, 64)
		var secondsModulus = int(vmUptimeInt) % 60.0

		var minutesSince = (float64(vmUptimeInt) - float64(secondsModulus)) / 60.0
		var minutesModulus = int(minutesSince) % 60.0

		var hoursSince = (minutesSince - float64(minutesModulus)) / 60
		var hoursModulus = int(hoursSince) % 24

		var daysSince = (int(hoursSince) - hoursModulus) / 24

		if finalResult != "-" {
			finalResult = strconv.Itoa(daysSince) + "d "
			finalResult = finalResult + strconv.Itoa(hoursModulus) + "h "
			finalResult = finalResult + strconv.Itoa(minutesModulus) + "m "
			finalResult = finalResult + strconv.Itoa(secondsModulus) + "s"
		}
	}
	return finalResult
}

func getOsDiskFullSize(vmName string) string {
	// fmt.Println("VM Name: " + vmName)
	var filePath = getVmFolder(vmName) + "/disk0.img"
	var osDiskLs string
	var osDiskLsArg1 = "ls"
	var osDiskLsArg2 = "-ahl"
	var osDiskLsArg3 = filePath

	var cmd = exec.Command(osDiskLsArg1, osDiskLsArg2, osDiskLsArg3)
	var stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func getOsDiskFullSize: There has been an error:", err)
		os.Exit(1)
	} else {
		osDiskLs = string(stdout)
	}

	var osDiskLsList []string
	for _, i := range strings.Split(osDiskLs, " ") {
		if len(i) > 1 {
			osDiskLsList = append(osDiskLsList, i)
			// fmt.Println(n, i)
		}
	}
	osDiskLs = osDiskLsList[3]
	// fmt.Println(osDiskLs)go
	return osDiskLs
}

func getOsDiskFree(vmName string) string {
	// fmt.Println("VM Name: " + vmName)
	var filePath = getVmFolder(vmName) + "/disk0.img"
	var osDiskDu string
	var osDiskDuArg1 = "du"
	var osDiskDuArg2 = "-h"
	var osDiskDuArg3 = filePath

	var cmd = exec.Command(osDiskDuArg1, osDiskDuArg2, osDiskDuArg3)
	var stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func getOsDiskFullSize: There has been an error:", err)
		os.Exit(1)
	} else {
		osDiskDu = string(stdout)
	}

	var osDiskDuList []string
	for _, i := range strings.Split(osDiskDu, "/") {
		if len(i) > 1 {
			osDiskDuList = append(osDiskDuList, i)
		}
	}

	// for n, i := range osDiskDuList {
	// fmt.Println(n, i)
	// }
	osDiskDu = osDiskDuList[0]
	osDiskDu = strings.ReplaceAll(osDiskDu, "\t", "")
	osDiskDu = strings.ReplaceAll(osDiskDu, " ", "")
	// fmt.Println(osDiskDu)
	return osDiskDu
}

type VmConfigStruct struct {
	CPUSockets string `json:"cpu_sockets"`
	CPUCores   string `json:"cpu_cores"`
	Memory     string `json:"memory"`
	Loader     string `json:"loader"`
	LiveStatus string `json:"live_status"`
	OsType     string `json:"os_type"`
	OsComment  string `json:"os_comment"`
	Owner      string `json:"owner"`
	ParentHost string `json:"parent_host"`
	Networks   []struct {
		NetworkAdaptorType string `json:"network_adaptor_type"`
		NetworkBridge      string `json:"network_bridge"`
		NetworkMac         string `json:"network_mac"`
		IPAddress          string `json:"ip_address"`
		Comment            string `json:"comment"`
	} `json:"networks"`
	Disks []struct {
		DiskType     string `json:"disk_type"`
		DiskLocation string `json:"disk_location"`
		DiskImage    string `json:"disk_image"`
		Comment      string `json:"comment"`
	} `json:"disks"`
	IncludeHostwideSSHKeys bool `json:"include_hostwide_ssh_keys"`
	VMSSHKeys              []struct {
		KeyOwner string `json:"key_owner"`
		KeyValue string `json:"key_value"`
		Comment  string `json:"comment"`
	} `json:"vm_ssh_keys"`
	VncPort     string `json:"vnc_port"`
	VncPassword string `json:"vnc_password"`
	Description string `json:"description"`
}
