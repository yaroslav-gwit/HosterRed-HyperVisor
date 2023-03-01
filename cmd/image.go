package cmd

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path"

	"facette.io/natsort"
	"github.com/schollz/progressbar/v3"
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
	var vmImageMap map[string][]map[string][]string
	err = json.Unmarshal(body, &vmImageMap)
	if err != nil {
		return err
	}
	var imageList []string
	for _, v := range vmImageMap["vm_images"] {
		for key, vv := range v {
			if key == osType {
				imageList = vv
				natsort.Sort(imageList)
			}
		}
	}
	if len(imageList) > 0 {
		vmImage := imageList[len(imageList)-1]
		vmImageFullLink := hostConfig.ImageServer + "images/" + vmImage
		req, err := http.NewRequest("GET", vmImageFullLink, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		f, _ := os.OpenFile("/tmp/"+osType+".zip", os.O_CREATE|os.O_WRONLY, 0600)
		defer f.Close()

		bar := progressbar.DefaultBytes(
			resp.ContentLength,
			"downloading",
		)
		io.Copy(io.MultiWriter(f, bar), resp.Body)
	} else {
		return errors.New("sorry, could not find the image")
	}

	return nil
}
