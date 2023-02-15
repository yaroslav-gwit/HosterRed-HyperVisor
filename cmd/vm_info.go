package cmd

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	jsonVmInfo       bool
	jsonPrettyVmInfo bool
	// vmInfoVmName     string

	vmInfoCmd = &cobra.Command{
		Use:   "info [vm name]",
		Short: "Print out the VM Info",
		Long:  `Print out the VM Info.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// fmt.Println(args)
			printVmInfo(args[0])
		},
	}
)

func printVmInfo(vmName string) {
	vmInfo, err := getVmInfo(vmName)
	if err != nil {
		log.Fatal(err)
	}
	if jsonPrettyVmInfo {
		jsonPretty, err := json.MarshalIndent(vmInfo, "", "   ")
		if err != nil {
			log.Fatal(err)
		}
		println(string(jsonPretty))
	} else {
		jsonOutput, err := json.Marshal(vmInfo)
		if err != nil {
			log.Fatal(err)
		}
		println(string(jsonOutput))
	}
}

type vmInfoStruct struct {
	VmName             string `json:"vm_name,omitempty"`
	MainIpAddress      string `json:"main_ip_address,omitempty"`
	VmStatusLive       bool   `json:"vm_status_live,omitempty"`
	VmStatusEncrypted  bool   `json:"vm_status_encrypted,omitempty"`
	VmStatusProduction bool   `json:"vm_status_production,omitempty"`
	CpuSockets         int    `json:"cpu_sockets,omitempty"`
	CpuCores           int    `json:"cpu_cores,omitempty"`
	RamAmount          string `json:"ram_amount,omitempty"`
	VncPort            int    `json:"vnc_port,omitempty"`
	VncPassword        string `json:"vnc_password,omitempty"`
	OsType             string `json:"os_type,omitempty"`
	VmUptime           string `json:"vm_uptime,omitempty"`
	VmDescription      string `json:"vm_description,omitempty"`
	ParentHost         string `json:"parent_host,omitempty"`
	Uptime             string `json:"uptime,omitempty"`
	OsDiskTotal        string `json:"os_disk_total,omitempty"`
	OsDiskUsed         string `json:"os_disk_used,omitempty"`
}

func getVmInfo(vmName string) (vmInfoStruct, error) {
	var vmInfoVar = vmInfoStruct{}
	vmInfoVar.VmName = vmName

	allVms := getAllVms()
	if slices.Contains(allVms, vmName) {
		_ = true
	} else {
		return vmInfoStruct{}, errors.New("VM is not found in the system")
	}

	wg.Add(1)
	go func() { defer wg.Done(); vmInfoVar.ParentHost = GetHostName() }()

	wg.Add(1)
	go func() { defer wg.Done(); vmInfoVar.VmStatusEncrypted = encryptionCheck(vmName) }()

	wg.Add(1)
	go func() { defer wg.Done(); vmInfoVar.OsDiskTotal = getOsDiskFullSize(vmName) }()

	wg.Add(1)
	go func() { defer wg.Done(); vmInfoVar.OsDiskUsed = getOsDiskUsed(vmName) }()

	wg.Add(1)
	go func() { defer wg.Done(); vmInfoVar.Uptime = getVmUptimeNew(vmName) }()

	wg.Add(1)
	go func() {
		defer wg.Done()
		vmConfigVar := vmConfig(vmName)
		vmInfoVar.MainIpAddress = vmConfigVar.Networks[0].IPAddress
	}()

	wg.Wait()
	return vmInfoVar, nil
}
