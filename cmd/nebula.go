package cmd

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	nebulaCmd = &cobra.Command{
		Use:   "nebula",
		Short: "Nebula network service manager",
		Long:  `Nebula network service manager`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

var (
	nebulaInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize Nebula on this node",
		Long:  `Initialize Nebula on this node (requires valid Nebula JSON config file)`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

var (
	nebulaShowLogCmd = &cobra.Command{
		Use:   "show-log",
		Short: "Use `tail -f` to display Nebula's live log",
		Long:  `Use "tail -f" to display Nebula's live log`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

var (
	nebulaServiceStart  bool
	nebulaServiceStop   bool
	nebulaServiceReload bool

	nebulaServiceCmd = &cobra.Command{
		Use:   "service",
		Short: "Start, stop, or reload Nebula process",
		Long:  `Start, stop, or reload Nebula process`,
		Run: func(cmd *cobra.Command, args []string) {
			// cmd.Help()
			reloadNebulaService()
		},
	}
)

var (
	nebulaUpdateBinary bool
	nebulaUpdateConfig bool

	nebulaUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Download the latest changes from Nebula Control Plane API server",
		Long:  `Download the latest changes from Nebula Control Plane API server`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

// const nebulaServiceFolder = "/opt/nebula_new/"
const nebulaServiceFolder = "/opt/nebula/"

func reloadNebulaService() error {
	reMatchLocation := regexp.MustCompile(`.*` + nebulaServiceFolder + `nebula.*`)
	reMatchSpace := regexp.MustCompile(`\s+`)
	pgrepOut, err := exec.Command("pgrep", "-lf", "nebula").CombinedOutput()
	if err != nil {
		return err
	}

	nebulaPid := ""
	for _, v := range strings.Split(string(pgrepOut), "\n") {
		if reMatchLocation.MatchString(v) {
			nebulaPid = reMatchSpace.Split(v, -1)[0]
		}
	}

	if len(nebulaPid) > 0 {
		fmt.Println(nebulaPid)
	} else {
		fmt.Println("Service is not running!")
	}

	return nil
}
