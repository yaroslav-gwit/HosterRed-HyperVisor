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

type VmImage struct {
	Almalinux8   []string `json:"almalinux8,omitempty"`
	Rockylinux8  []string `json:"rockylinux8,omitempty"`
	Ubuntu2004   []string `json:"ubuntu2004,omitempty"`
	Debian11     []string `json:"debian11,omitempty"`
	FreeBsd13Zfs []string `json:"freebsd13zfs,omitempty"`
	FreeBsd13Ufs []string `json:"freebsd13ufs,omitempty"`
	Windows10    []string `json:"windows10,omitempty"`
}

type VmImages struct {
	VmImages []VmImage `json:"vm_images"`
}

func imageDownload(osType string, force bool) error {
	var vmImages VmImages

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
	err = json.Unmarshal(body, &vmImages)
	if err != nil {
		return err
	}

	var imageList []string
	for _, v := range vmImages.VmImages {
		switch osType {
		case "almalinux8":
			imageList = v.Almalinux8
			natsort.Sort(imageList)
		case "rockylinux8":
			imageList = v.Rockylinux8
			natsort.Sort(imageList)
		case "ubuntu2004":
			imageList = v.Ubuntu2004
			natsort.Sort(imageList)
		case "debian11":
			imageList = v.Debian11
			natsort.Sort(imageList)
		case "freebsd13zfs":
			imageList = v.FreeBsd13Zfs
			natsort.Sort(imageList)
		case "freebsd13ufs":
			imageList = v.FreeBsd13Ufs
			natsort.Sort(imageList)
		case "windows10":
			imageList = v.Windows10
			natsort.Sort(imageList)
		default:
			imageList = v.Debian11
			natsort.Sort(imageList)
		}
	}

	if len(imageList) > 0 {
		println(imageList[len(imageList)-1])
		println("Full image link: " + hostConfig.ImageServer + "images/" + imageList[len(imageList)-1])
	} else {
		println("Image list is empty, sorry")
	}
	println(osType)
	println(imageList)

	return nil
}
