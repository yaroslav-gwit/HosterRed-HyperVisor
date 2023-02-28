package cmd

import (
	"hoster/emojlog"
	"log"

	"github.com/spf13/cobra"
)

var (
	replicationEndpointAll string
	endpointSshPortAll     int
	sshKeyLocationAll      string

	vmReplicateAllCmd = &cobra.Command{
		Use:   "replicate-all",
		Short: "Replicate all live and production VMs to a backup node",
		Long:  `Replicate all live and production VMs to a backup node.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(replicationEndpoint) < 1 {
				log.Fatal("Please specify an endpoint!")
			}
			replicateAllProdVms(replicationEndpointAll, endpointSshPortAll, sshKeyLocationAll)
		},
	}
)

func replicateAllProdVms(replicationEndpoint string, endpointSshPort int, sshKeyLocation string) {
	for _, v := range getAllVms() {
		vmConfigVar := vmConfig(v)
		if vmConfigVar.ParentHost != GetHostName() {
			continue
		}
		if !vmLiveCheck(v) {
			continue
		}
		err := replicateVm(v, replicationEndpoint, endpointSshPort, sshKeyLocation)
		if err != nil {
			emojlog.PrintLogMessage("Replication failed for a VM: "+v+" || Exact error: "+err.Error(), emojlog.Error)
		}
	}
}
