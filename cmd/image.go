package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"

	"facette.io/natsort"
	"github.com/spf13/cobra"
)

var (
	imageCmd = &cobra.Command{
		Use:   "image",
		Short: "Image and template (.raw) related operations",
		Long:  `Image and template (.raw) related operations`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

var (
	imageOsType        string
	imageForceDownload bool

	imageDownloadCmd = &cobra.Command{
		Use:   "download",
		Short: "Download an image from the public or private repo",
		Long:  `Download an image from the public or private repo`,
		Run: func(cmd *cobra.Command, args []string) {
			imageDownload(imageOsType, imageForceDownload)
		},
	}
)

func imageDownload(osType string, force bool) error {
	// Host config read/parse
	hostConfig := HostConfig{}
	// JSON config file location
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	hostConfigFile := path.Dir(execPath) + "/config_files/host_config.json"
	// Read the JSON file
	data, err := os.ReadFile(hostConfigFile)
	if err != nil {
		return err
	}
	// Unmarshal the JSON data into a slice of Network structs
	err = json.Unmarshal(data, &hostConfig)
	if err != nil {
		return err
	}

	// Parse website response
	resp, err := http.Get(hostConfig.ImageServer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var m map[string][]map[string][]string
	err = json.Unmarshal(body, &m)
	if err != nil {
		panic(err)
	}

	var imageList []string
	for _, v := range m["vm_images"] {
		for key, vv := range v {
			if key == osType {
				// fmt.Println(k, v)
				imageList = vv
				natsort.Sort(imageList)
			}
		}
	}

	if len(imageList) > 0 {
		println(imageList[len(imageList)-1])
		println("Full image link: " + hostConfig.ImageServer + "images/" + imageList[len(imageList)-1])
	} else {
		println("Image list is empty, sorry")
	}

	return nil
}
