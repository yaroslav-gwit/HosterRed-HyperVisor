package cmd

import (
	"encoding/json"
	"fmt"
	"log"
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
	jsonHostInfoOutput       bool
	jsonPrettyHostInfoOutput bool

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
	if jsonPrettyHostInfoOutput {
		jsonOutputVar := jsonOutputHostInfo()
		jsonData, err := json.MarshalIndent(jsonOutputVar, "", "   ")
		if err != nil {
			log.Fatal("Function error: HostMain:", err)
		}
		fmt.Println(string(jsonData))
	} else if jsonHostInfoOutput {
		jsonOutputVar := jsonOutputHostInfo()
		jsonData, err := json.Marshal(jsonOutputVar)
		if err != nil {
			log.Fatal("Function error: HostMain:", err)
		}
		fmt.Println(string(jsonData))
	} else {
		var tHostname string
		var tLiveVms string
		var tSystemUptime string
		var tSystemRam = ramResponse{}
		var tSwapInfo swapInfoStruct
		var tArcSize string
		var tFreeZfsSpace string
		var tZrootStatus string

		var wg = &sync.WaitGroup{}
		var err error
		wg.Add(8)
		go func() { defer wg.Done(); tHostname = GetHostName() }()
		go func() { defer wg.Done(); tLiveVms = getNumberOfRunningVms() }()
		go func() { defer wg.Done(); tSystemUptime = getSystemUptime() }()
		go func() { defer wg.Done(); tSystemRam = getHostRam() }()
		go func() { defer wg.Done(); tArcSize = getArcSize() }()
		go func() { defer wg.Done(); tFreeZfsSpace = getFreeZfsSpace() }()
		go func() { defer wg.Done(); tZrootStatus = getZrootStatus() }()
		go func() {
			defer wg.Done()
			tSwapInfo, err = getSwapInfo()
			if err != nil {
				log.Fatal(err)
			}
		}()
		wg.Wait()

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
			"RAM (Used/Total)",
			"SWAP (Used/Total)",
			"ZFS ARC Size",
			"Zroot (Used/Total)",
			"Zroot Status",
		)

		t.AddRow(tHostname,
			tLiveVms,
			tSystemUptime,
			tSystemRam.used+"/"+tSystemRam.all,
			tSwapInfo.used+"/"+tSwapInfo.total,
			tArcSize,
			tFreeZfsSpace,
			tZrootStatus,
		)

		t.Render()
	}
}

type jsonOutputHostInfoStruct struct {
	Hostname     string `json:"hostname"`
	LiveVms      string `json:"live_vms"`
	SystemUptime string `json:"system_uptime"`
	RamTotal     string `json:"ram_total"`
	RamFree      string `json:"ram_free"`
	RamUsed      string `json:"ram_used"`
	SwapTotal    string `json:"swap_total"`
	SwapUsed     string `json:"swap_used"`
	SwapFree     string `json:"swap_free"`
	ArcSize      string `json:"zfs_acr_size"`
	ZrootFree    string `json:"zroot_free"`
	ZrootStatus  string `json:"zroot_status"`
}

func jsonOutputHostInfo() jsonOutputHostInfoStruct {
	var tHostname string
	var tLiveVms string
	var tSystemUptime string
	var tSystemRam = ramResponse{}
	var tSwapInfo swapInfoStruct
	var tArcSize string
	var tFreeZfsSpace string
	var tZrootStatus string

	var wg = &sync.WaitGroup{}
	var err error
	wg.Add(8)
	go func() { defer wg.Done(); tHostname = GetHostName() }()
	go func() { defer wg.Done(); tLiveVms = getNumberOfRunningVms() }()
	go func() { defer wg.Done(); tSystemUptime = getSystemUptime() }()
	go func() { defer wg.Done(); tSystemRam = getHostRam() }()
	go func() { defer wg.Done(); tArcSize = getArcSize() }()
	go func() { defer wg.Done(); tFreeZfsSpace = getFreeZfsSpace() }()
	go func() { defer wg.Done(); tZrootStatus = getZrootStatus() }()

	go func() {
		defer wg.Done()
		tSwapInfo, err = getSwapInfo()
		if err != nil {
			log.Fatal(err)
		}
	}()
	wg.Wait()

	jsonOutputVar := jsonOutputHostInfoStruct{}
	jsonOutputVar.Hostname = tHostname
	jsonOutputVar.LiveVms = tLiveVms
	jsonOutputVar.SystemUptime = tSystemUptime
	jsonOutputVar.RamTotal = tSystemRam.all
	jsonOutputVar.RamFree = tSystemRam.free
	jsonOutputVar.RamUsed = tSystemRam.used
	jsonOutputVar.SwapTotal = tSwapInfo.total
	jsonOutputVar.SwapUsed = tSwapInfo.used
	jsonOutputVar.SwapFree = tSwapInfo.free
	jsonOutputVar.ArcSize = tArcSize
	jsonOutputVar.ZrootFree = tFreeZfsSpace
	jsonOutputVar.ZrootStatus = tZrootStatus

	return jsonOutputVar
}

type ramResponse struct {
	free string
	used string
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
	vFreeCountArg1 := "sysctl"
	vFreeCountArg2 := "-nq"
	vFreeCountArg3 := "vm.stats.vm.v_free_count"

	cmd := exec.Command(vFreeCountArg1, vFreeCountArg2, vFreeCountArg3)
	stdout, err := cmd.Output()
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
	realMemArg1 := "sysctl"
	realMemArg2 := "-nq"
	realMemArg3 := "hw.realmem"

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

	vFreeCountInt, _ := strconv.Atoi(vFreeCount)
	hwPagesizeInt, _ := strconv.Atoi(hwPagesize)
	realMemInt, _ := strconv.Atoi(realMem)

	finalResultFree := vFreeCountInt * hwPagesizeInt
	finalResultReal := realMemInt
	finalResultUsed := (finalResultReal - finalResultFree)

	ramResponseVar := ramResponse{}
	ramResponseVar.free = ByteConversion(finalResultFree)
	ramResponseVar.used = ByteConversion(finalResultUsed)
	ramResponseVar.all = ByteConversion(finalResultReal)

	return ramResponseVar
}

func getArcSize() string {
	// GET SYSCTL "vm.stats.vm.v_free_count" AND RETURN THE VALUE
	var arcSize string
	arcSizeArg1 := "sysctl"
	arcSizeArg2 := "-nq"
	arcSizeArg3 := "kstat.zfs.misc.arcstats.size"

	cmd := exec.Command(arcSizeArg1, arcSizeArg2, arcSizeArg3)
	stdout, err := cmd.Output()
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

	acrSizeInt, _ := strconv.Atoi(arcSizeList[0])
	return ByteConversion(acrSizeInt)
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

type swapInfoStruct struct {
	total string
	used  string
	free  string
}

func getSwapInfo() (swapInfoStruct, error) {
	swapInfoVar := swapInfoStruct{}
	stdout, stderr := exec.Command("swapinfo").Output()
	if stderr != nil {
		return swapInfoStruct{}, stderr
	}

	reSplitSpace := regexp.MustCompile(`\s+`)
	var swapInfoList []string
	for _, v := range reSplitSpace.Split(string(stdout), -1) {
		if len(v) > 1 {
			swapInfoList = append(swapInfoList, v)
		}
	}
	swapTotalBytes, _ := strconv.Atoi(swapInfoList[6])
	swapTotalBytes = swapTotalBytes * 1024
	swapUsedBytes, _ := strconv.Atoi(swapInfoList[7])
	swapUsedBytes = swapUsedBytes * 1024
	swapFreeBytes, _ := strconv.Atoi(swapInfoList[8])
	swapFreeBytes = swapFreeBytes * 1024

	swapInfoVar.total = ByteConversion(swapTotalBytes)
	swapInfoVar.free = ByteConversion(swapFreeBytes)
	swapInfoVar.used = ByteConversion(swapUsedBytes)

	return swapInfoVar, nil
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
