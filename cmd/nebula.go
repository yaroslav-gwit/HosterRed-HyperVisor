package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"hoster/emojlog"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
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
			err = downloadNebulaCerts()
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

type NebulaClusterConfig struct {
	ClusterName   string `json:"cluster_name"`
	ClusterID     string `json:"cluster_id"`
	HostID        string `json:"host_id"`
	NatPunch      string `json:"nat_punch"`
	ListenAddress string `json:"listen_address"`
	ListenPort    string `json:"listen_port"`
	MTU           string `json:"mtu"`
	UseRelays     string `json:"use_relays"`
	APIServer     string `json:"api_server"`
}

func readNebulaClusterConfig() (NebulaClusterConfig, error) {
	execPath, err := os.Executable()
	if err != nil {
		return NebulaClusterConfig{}, err
	}
	nebulaClusterConfigFile := path.Dir(execPath) + "/config_files/nebula_config.json"
	// read the json file from disk
	data, err := os.ReadFile(nebulaClusterConfigFile)
	if err != nil {
		return NebulaClusterConfig{}, err
	}

	// unmarshal the json data into a Config struct
	var config NebulaClusterConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return NebulaClusterConfig{}, err
	}

	return NebulaClusterConfig{}, nil
}

func downloadNebulaConfig() error {
	c, err := readNebulaClusterConfig()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", "https://"+c.APIServer+"/get_config?cluster_name="+c.ClusterName+"&cluster_id="+c.ClusterID+"&host_name="+GetHostName()+"&host_id="+c.HostID+"&nat_punch="+c.NatPunch+"&listen_host="+c.ListenAddress+"&listen_port="+c.ListenPort+"&mtu="+c.MTU+"&use_relays="+c.UseRelays, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(body))

	return nil
}

func downloadNebulaCerts() error {
	req, err := http.NewRequest("GET", "https://fastapi-test.yari.pw/get_certs?cluster_name=GWIT&cluster_id=ocK7U4Xd&host_name=hoster-test-0101&host_id=UqKvh5YU", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(body))

	return nil
}
