package cmd

import (
	"encoding/json"
	"log"

	"github.com/spf13/cobra"
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
	vmInfo := getVmInfo(vmName)
	if jsonPrettyVmInfo {
		jsonPretty, err := json.MarshalIndent(vmInfo, "", "   ")
		if err != nil {
			log.Fatal(err)
		}
		println(string(jsonPretty))
	} else if jsonVmInfo {
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
}

func getVmInfo(vmName string) vmInfoStruct {
	var vmInfoVar = vmInfoStruct{}
	vmInfoVar.VmName = vmName
	wg.Add(1)
	go func() { defer wg.Done(); vmInfoVar.ParentHost = GetHostName() }()
	go func() { defer wg.Done(); vmInfoVar.VmStatusEncrypted = encryptionCheck(vmName) }()
	wg.Wait()

	return vmInfoVar
}
