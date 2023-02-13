package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool

	hostCmd = &cobra.Command{
		Use:   "host",
		Short: "Host related operations",
		Long:  `Host related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			HostMain()
		},
	}
)

func HostMain() {
	// USE GOLANG SYNC LIB TO GET ALL THE DATA CONCURRENTLY
	var tHostname string
	var tLiveVms string
	var tSystemUptime string
	var tSystemRam = ramResponse{}
	var tFreeSwap string
	var tAllSwap string
	var tArcSize string
	var tFreeZfsSpace string
	var tZrootStatus string

	var wg = &sync.WaitGroup{}
	wg.Add(9)
	go func() { defer wg.Done(); tHostname = GetHostName() }()
	go func() { defer wg.Done(); tLiveVms = getNumberOfRunningVms() }()
	go func() { defer wg.Done(); tSystemUptime = getSystemUptime() }()
	go func() { defer wg.Done(); tSystemRam = getHostRam() }()
	go func() { defer wg.Done(); tFreeSwap = getFreeSwapSpace() }()
	go func() { defer wg.Done(); tAllSwap = getAllSwapSpace() }()
	go func() { defer wg.Done(); tArcSize = getArcSize() }()
	go func() { defer wg.Done(); tFreeZfsSpace = getFreeZfsSpace() }()
	go func() { defer wg.Done(); tZrootStatus = getZrootStatus() }()
	wg.Wait()

	if jsonOutput {
		var jsonOutputVar = jsonOutputHostInfo{}
		jsonOutputVar.Hostname = tHostname
		jsonOutputVar.LiveVms = tLiveVms
		jsonOutputVar.SystemUptime = tSystemUptime
		jsonOutputVar.FreeRAM = tSystemRam.free
		jsonOutputVar.AllSystemRAM = tSystemRam.all
		jsonOutputVar.FreeSwap = tFreeSwap
		jsonOutputVar.ArcSize = tArcSize
		jsonOutputVar.ZrootFree = tFreeZfsSpace
		jsonOutputVar.ZrootStatus = tZrootStatus

		var jsonData, _ = json.MarshalIndent(jsonOutputVar, "", "   ")
		fmt.Println(string(jsonData))
	} else {
		t := table.New(os.Stdout)
		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)

		t.SetAlignment(
			table.AlignLeft,   // Hostname
			table.AlignCenter, // Live VMs
			table.AlignCenter, // System Uptime
			table.AlignCenter, // RAM
			table.AlignCenter, // SWAP
			table.AlignCenter, // ARC Size
			table.AlignCenter, // Zroot space
			table.AlignCenter, // Zpool status
		)

		t.SetHeaders("Host information")
		t.SetHeaderColSpans(0, 8)
		t.AddHeaders(
			"Hostname",
			"Live VMs",
			"System Uptime",
			"RAM (Free)",
			"SWAP (Free)",
			"ARC Size",
			"Zroot Space Free",
			"Zroot Pool Status",
		)

		t.AddRow(tHostname,
			tLiveVms,
			tSystemUptime,
			tSystemRam.free+"/"+tSystemRam.all,
			tFreeSwap+"/"+tAllSwap,
			tArcSize,
			tFreeZfsSpace,
			tZrootStatus,
		)

		t.Render()
	}
}

type jsonOutputHostInfo struct {
	Hostname     string `json:"hostname"`
	LiveVms      string `json:"live_vms"`
	SystemUptime string `json:"system_uptime"`
	FreeRAM      string `json:"free_ram"`
	AllSystemRAM string `json:"all_system_ram"`
	FreeSwap     string `json:"free_swap"`
	ArcSize      string `json:"zfs_acr_size"`
	ZrootFree    string `json:"zroot_free"`
	ZrootStatus  string `json:"zroot_status"`
}

type ramResponse struct {
	free string
	all  string
}

// type swapResponse struct {
// 	free string
// 	all  string
// }

// type zrootUsageResponse struct {
// 	free string
// 	all  string
// }

func getHostRam() ramResponse {
	// GET SYSCTL "vm.stats.vm.v_free_count" AND RETURN THE VALUE
	var vFreeCount string
	var vFreeCountArg1 = "sysctl"
	var vFreeCountArg2 = "-nq"
	var vFreeCountArg3 = "vm.stats.vm.v_free_count"

	var cmd = exec.Command(vFreeCountArg1, vFreeCountArg2, vFreeCountArg3)
	var stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func freeRam/vFreeCount: There has been an error:", err)
		os.Exit(1)
	} else {
		vFreeCount = string(stdout)
	}

	var vFreeCountList []string
	for _, i := range strings.Split(vFreeCount, "\n") {
		if len(i) > 1 {
			vFreeCountList = append(vFreeCountList, i)
		}
	}
	vFreeCount = vFreeCountList[0]

	var realMem string
	var realMemArg1 = "sysctl"
	var realMemArg2 = "-nq"
	var realMemArg3 = "hw.realmem"

	cmd = exec.Command(realMemArg1, realMemArg2, realMemArg3)
	stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func freeRam/vFreeCount: There has been an error:", err)
		os.Exit(1)
	} else {
		realMem = string(stdout)
	}

	var realMemList []string
	for _, i := range strings.Split(realMem, "\n") {
		if len(i) > 1 {
			realMemList = append(realMemList, i)
		}
	}
	realMem = realMemList[0]

	// GET SYSCTL "hw.pagesize" AND RETURN THE VALUE
	var hwPagesize string
	var hwPagesizeArg1 = "sysctl"
	var hwPagesizeArg2 = "-nq"
	var hwPagesizeArg3 = "hw.pagesize"
	cmd = exec.Command(hwPagesizeArg1, hwPagesizeArg2, hwPagesizeArg3)
	stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func freeRam/hwPagesize: There has been an error:", err)
		os.Exit(1)
	} else {
		hwPagesize = string(stdout)
	}
	var hwPagesizeList []string
	for _, i := range strings.Split(hwPagesize, "\n") {
		if len(i) > 1 {
			hwPagesizeList = append(hwPagesizeList, i)
		}
	}
	hwPagesize = hwPagesizeList[0]

	var vFreeCountInt, _ = strconv.Atoi(vFreeCount)
	var hwPagesizeInt, _ = strconv.Atoi(hwPagesize)
	var realMemInt, _ = strconv.Atoi(realMem)

	var finalResultFree = vFreeCountInt * hwPagesizeInt
	// var finalResultReal = realMemInt * hwPagesizeInt
	var finalResultReal = realMemInt
	var finalResultFreeNotBytes = ByteConversion(finalResultFree)
	var finalResultRealNotBytes = ByteConversion(finalResultReal)

	var ramResponseVar = ramResponse{}
	ramResponseVar.free = finalResultFreeNotBytes
	ramResponseVar.all = finalResultRealNotBytes

	return ramResponseVar
}

func getArcSize() string {
	// GET SYSCTL "vm.stats.vm.v_free_count" AND RETURN THE VALUE
	var arcSize string
	var arcSizeArg1 = "sysctl"
	var arcSizeArg2 = "-nq"
	var arcSizeArg3 = "kstat.zfs.misc.arcstats.size"

	var cmd = exec.Command(arcSizeArg1, arcSizeArg2, arcSizeArg3)
	var stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func getArcSize/arcSize: There has been an error:", err)
		os.Exit(1)
	} else {
		arcSize = string(stdout)
	}

	var arcSizeList []string
	for _, i := range strings.Split(arcSize, "\n") {
		if len(i) > 1 {
			arcSizeList = append(arcSizeList, i)
		}
	}
	arcSize = arcSizeList[0]

	var acrSizeInt, _ = strconv.Atoi(arcSize)
	var finalResult = ByteConversion(acrSizeInt)

	return finalResult
}

func getNumberOfRunningVms() string {
	files, err := os.ReadDir("/dev/vmm/")
	var finalResult string
	if err != nil {
		// fmt.Println("funcError getNumberOfRunningVms: " + err.Error())
		// os.Exit(1)
		finalResult = "0"
	} else {
		finalResult = strconv.Itoa(len(files))

	}

	return finalResult
}

func getZrootStatus() string {
	var zrootStatus string
	var zrootStatusArg1 = "zpool"
	var zrootStatusArg2 = "status"
	var zrootStatusArg3 = "zroot"

	var cmd = exec.Command(zrootStatusArg1, zrootStatusArg2, zrootStatusArg3)
	var stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func getZrootStatus/zrootStatus: There has been an error:", err)
		os.Exit(1)
	} else {
		zrootStatus = string(stdout)
	}
	var zrootStatusList []string
	for _, i := range strings.Split(zrootStatus, "\n") {
		if len(i) > 1 {
			zrootStatusList = append(zrootStatusList, i)
		}
	}

	var r, _ = regexp.Compile(".*state:.*")
	for _, i := range zrootStatusList {
		var reMatch = r.MatchString(i)
		if reMatch {
			zrootStatus = i
		}
	}

	zrootStatus = strings.Replace(zrootStatus, "state:", "", -1)
	zrootStatus = strings.Replace(zrootStatus, " ", "", -1)
	if zrootStatus == "ONLINE" {
		zrootStatus = "Online"
	} else {
		zrootStatus = "Problem!"
	}

	var finalResult = zrootStatus

	return finalResult
}

func getFreeZfsSpace() string {
	var zrootFree string
	var zrootFreeArg1 = "zfs"
	var zrootFreeArg2 = "list"
	var zrootFreeArg3 = "zroot"

	var cmd = exec.Command(zrootFreeArg1, zrootFreeArg2, zrootFreeArg3)
	var stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func getFreeZfsSpace/zrootFree: There has been an error:", err)
		os.Exit(1)
	} else {
		zrootFree = string(stdout)
	}
	var zrootFreeList []string
	for _, i := range strings.Split(zrootFree, " ") {
		if len(i) > 1 {
			zrootFreeList = append(zrootFreeList, i)
		}
	}

	zrootFree = zrootFreeList[6]
	zrootFree = strings.Replace(zrootFree, " ", "", -1)

	var finalResult = zrootFree
	return finalResult
}

func getFreeSwapSpace() string {
	var swapFree string
	var swapFreeArg1 = "swapinfo"

	var cmd = exec.Command(swapFreeArg1)
	var stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func getFreeZfsSpace/zrootFree: There has been an error:", err)
		os.Exit(1)
	} else {
		swapFree = string(stdout)
	}
	var swapFreeList []string
	for _, i := range strings.Split(swapFree, " ") {
		if len(i) > 1 {
			swapFreeList = append(swapFreeList, i)
		}
	}

	// CONVERT KB TO A DIFFERENT BYTE TYPE IF NECESSARY
	swapFree = swapFreeList[7]
	var swapFreeBytes, _ = strconv.Atoi(swapFree)
	var byteType = "K"
	if len(swapFree) > 3 {
		swapFreeBytes = swapFreeBytes / 1024
		byteType = "M"
	}

	var swapFreeGb = 0.0
	if swapFreeBytes > 1024 {
		swapFreeGb = float64(swapFreeBytes) / 1024.0
		byteType = "G"
	}

	var finalResult string
	if swapFreeGb > 0.0 {
		finalResult = fmt.Sprintf("%.2f", swapFreeGb) + byteType
	} else {
		finalResult = strconv.Itoa(swapFreeBytes) + byteType
	}

	return finalResult
}

func getAllSwapSpace() string {
	var swapAll string
	var swapAllArg1 = "swapinfo"

	var cmd = exec.Command(swapAllArg1)
	var stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func getAllSwapSpace/swapAll: There has been an error:", err)
		os.Exit(1)
	} else {
		swapAll = string(stdout)
	}
	var swapAllList []string
	for _, i := range strings.Split(swapAll, " ") {
		if len(i) > 1 {
			swapAllList = append(swapAllList, i)
		}
	}

	// CONVERT KB TO A DIFFERENT BYTE TYPE IF NECESSARY
	swapAll = swapAllList[5]
	var swapAllBytes, _ = strconv.Atoi(swapAll)
	var byteType = "K"
	if len(swapAll) > 3 {
		swapAllBytes = swapAllBytes / 1024
		byteType = "M"
	}

	var swapAllGb = 0.0
	if swapAllBytes > 1024 {
		swapAllGb = float64(swapAllBytes) / 1024.0
		byteType = "G"
	}

	var finalResult string
	if swapAllGb > 0.0 {
		finalResult = fmt.Sprintf("%.2f", swapAllGb) + byteType
	} else {
		finalResult = strconv.Itoa(swapAllBytes) + byteType
	}

	return finalResult
}

func getSystemUptime() string {
	var systemUptime string
	var systemUptimeArg1 = "sysctl"
	var systemUptimeArg2 = "-nq"
	var systemUptimeArg3 = "kern.boottime"

	var cmd = exec.Command(systemUptimeArg1, systemUptimeArg2, systemUptimeArg3)
	var stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func getSystemUptime/systemUptime: There has been an error:", err)
		os.Exit(1)
	} else {
		systemUptime = string(stdout)
	}

	var systemUptimeList []string
	for _, i := range strings.Split(systemUptime, " ") {
		if len(i) > 1 {
			systemUptimeList = append(systemUptimeList, i)
		}
	}

	systemUptime = systemUptimeList[1]
	systemUptime = strings.Replace(systemUptime, ",", "", -1)
	systemUptime = strings.Replace(systemUptime, " ", "", -1)

	var systemUptimeInt, _ = strconv.ParseInt(systemUptime, 10, 64)
	var unixTime = time.Unix(systemUptimeInt, 0)

	var timeSince = time.Since(unixTime).Seconds()
	var secondsModulus = int(timeSince) % 60.0

	var minutesSince = (timeSince - float64(secondsModulus)) / 60.0
	var minutesModulus = int(minutesSince) % 60.0

	var hoursSince = (minutesSince - float64(minutesModulus)) / 60
	var hoursModulus = int(hoursSince) % 24

	var daysSince = (int(hoursSince) - hoursModulus) / 24

	var finalResult = strconv.Itoa(daysSince) + "d "
	finalResult = finalResult + strconv.Itoa(hoursModulus) + "h "
	finalResult = finalResult + strconv.Itoa(minutesModulus) + "m "
	finalResult = finalResult + strconv.Itoa(secondsModulus) + "s"

	return finalResult
}

func GetHostName() string {
	// GET SYSCTL "vm.stats.vm.v_free_count" AND RETURN THE VALUE
	var hostName string
	var hostNameArg1 = "sysctl"
	var hostNameArg2 = "-nq"
	var hostNameArg3 = "kern.hostname"

	var cmd = exec.Command(hostNameArg1, hostNameArg2, hostNameArg3)
	var stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func GetHostName/hostName: There has been an error:", err)
		os.Exit(1)
	} else {
		hostName = string(stdout)
	}

	var hostNameList []string
	for _, i := range strings.Split(hostName, "\n") {
		if len(i) > 1 {
			hostNameList = append(hostNameList, i)
		}
	}
	hostName = hostNameList[0]

	var finalResult = hostName

	return finalResult
}

func ByteConversion(bytes int) string {
	// SET TO KB
	var firstIteration = bytes / 1024
	var iterationTitle = "K"

	// SET TO MB
	if firstIteration > 1024 {
		firstIteration = firstIteration / 1024
		iterationTitle = "M"
	}

	// SET TO GB
	var firstIterationFloat = 0.0
	if firstIteration > 1024 {
		firstIterationFloat = float64(firstIteration) / 1024.0
		iterationTitle = "G"
	}

	// FORMAT THE OUTPUT
	var finalResult string
	if firstIterationFloat > 0.0 {
		finalResult = fmt.Sprintf("%.2f", firstIterationFloat) + iterationTitle
	} else {
		finalResult = strconv.Itoa(firstIteration) + iterationTitle
	}
	return finalResult
}
