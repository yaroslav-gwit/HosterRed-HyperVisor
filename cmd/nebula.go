package cmd

import (
	"bufio"
	"errors"
	"hoster/emojlog"
	"log"
	"os"
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
			err := tailNebulaLogFile()
			if err != nil {
				log.Fatal(err)
			}
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
			if nebulaServiceReload {
				reloadNebulaService()
			} else {
				cmd.Help()
			}
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
		return errors.New(string(pgrepOut))
	}

	nebulaPid := ""
	for _, v := range strings.Split(string(pgrepOut), "\n") {
		if reMatchLocation.MatchString(v) {
			nebulaPid = reMatchSpace.Split(v, -1)[0]
		}
	}

	if len(nebulaPid) > 0 {
		const nebulaStartSh = "(( nohup " + nebulaServiceFolder + "nebula -config " + nebulaServiceFolder + "config.yml 1>>" + nebulaServiceFolder + "log.txt 2>&1 )&)"
		const nebulaStartShLocation = "/tmp/nebula.sh"
		// Open nebulaStartShLocation for writing
		nebulaStartShFile, err := os.Create(nebulaStartShLocation)
		if err != nil {
			return err
		}
		defer nebulaStartShFile.Close()
		// Create a new writer
		writer := bufio.NewWriter(nebulaStartShFile)
		// Write a string to the file
		_, err = writer.WriteString(nebulaStartSh)
		if err != nil {
			return err
		}
		// Flush the writer to ensure all data has been written to the file
		err = writer.Flush()
		if err != nil {
			return err
		}
		err = os.Chmod(nebulaStartShLocation, os.FileMode(0600))
		if err != nil {
			return errors.New("error changing permissions: " + err.Error())
		}

		killOut, err := exec.Command("kill", "-SIGTERM", nebulaPid).CombinedOutput()
		if err != nil {
			return errors.New(string(killOut))
		}
		emojlog.PrintLogMessage("Stopped Nebula service using it's pid: "+nebulaPid, emojlog.Debug)
		nebulaStartErr := exec.Command("sh", nebulaStartShLocation).Start()
		if err != nil {
			return nebulaStartErr
		}
		emojlog.PrintLogMessage("Started new Nebula process", emojlog.Debug)
	} else {
		emojlog.PrintLogMessage("Service is not running", emojlog.Warning)
	}

	return nil
}

func tailNebulaLogFile() error {
	tailCmd := exec.Command("tail", "-f", nebulaServiceFolder, "log.txt")
	tailCmd.Stdin = os.Stdin
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr

	err := tailCmd.Run()
	if err != nil {
		return err
	}

	return nil
}
