package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"hoster/emojlog"
	"log"
	"net/http"
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
				err := reloadNebulaService()
				if err != nil {
					log.Fatal(err)
				}
			} else if nebulaServiceStart {
				err := startNebulaService()
				if err != nil {
					log.Fatal(err)
				}
			} else if nebulaServiceStop {
				err := stopNebulaService()
				if err != nil {
					log.Fatal(err)
				}
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
			// cmd.Help()
			err := downloadNebulaConfig()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

// const nebulaServiceFolder = "/opt/nebula_new/"
const nebulaServiceFolder = "/opt/nebula/"

func startNebulaService() error {
	reMatchLocation := regexp.MustCompile(`.*` + nebulaServiceFolder + `nebula.*`)
	reMatchSpace := regexp.MustCompile(`\s+`)
	pgrepOut, _ := exec.Command("pgrep", "-lf", "nebula").CombinedOutput()

	nebulaPid := ""
	for _, v := range strings.Split(string(pgrepOut), "\n") {
		if reMatchLocation.MatchString(v) {
			nebulaPid = reMatchSpace.Split(v, -1)[0]
		}
	}

	if len(nebulaPid) > 0 {
		return errors.New("service process for Nebula is already running")
	}

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

	nebulaStartErr := exec.Command("sh", nebulaStartShLocation).Start()
	if err != nil {
		return nebulaStartErr
	}
	emojlog.PrintLogMessage("Started new Nebula process", emojlog.Debug)

	return nil
}

func stopNebulaService() error {
	reMatchLocation := regexp.MustCompile(`.*` + nebulaServiceFolder + `nebula.*`)
	reMatchSpace := regexp.MustCompile(`\s+`)
	pgrepOut, _ := exec.Command("pgrep", "-lf", "nebula").CombinedOutput()

	nebulaPid := ""
	for _, v := range strings.Split(string(pgrepOut), "\n") {
		if reMatchLocation.MatchString(v) {
			nebulaPid = reMatchSpace.Split(v, -1)[0]
		}
	}

	if len(nebulaPid) < 1 {
		emojlog.PrintLogMessage("Nebula service is already dead: ", emojlog.Error)
		return errors.New("service is already dead")
	}

	killOut, err := exec.Command("kill", "-SIGTERM", nebulaPid).CombinedOutput()
	if err != nil {
		return errors.New(string(killOut))
	}
	emojlog.PrintLogMessage("Stopped Nebula service using it's pid: "+nebulaPid, emojlog.Debug)

	return nil
}

func reloadNebulaService() error {
	reMatchLocation := regexp.MustCompile(`.*` + nebulaServiceFolder + `nebula.*`)
	reMatchSpace := regexp.MustCompile(`\s+`)
	pgrepOut, _ := exec.Command("pgrep", "-lf", "nebula").CombinedOutput()

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
	tailCmd := exec.Command("tail", "-35", "-f", nebulaServiceFolder+"log.txt")
	tailCmd.Stdin = os.Stdin
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr

	err := tailCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func downloadNebulaConfig() error {
	req, err := http.NewRequest("GET", "https://fastapi-test.yari.pw/get_config?cluster_name=GWIT&cluster_id=ocK7U4Xd&host_name=hoster-test-0101&host_id=UqKvh5YU&nat_punch=true&listen_host=0.0.0.0&listen_port=14001&mtu=1300&use_relays=true", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println(resp.Body)

	return nil
}
