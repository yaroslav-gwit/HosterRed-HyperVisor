package cmd

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"facette.io/natsort"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
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
	imageDataset       string

	imageDownloadCmd = &cobra.Command{
		Use:   "download",
		Short: "Download an image from the public or private repo",
		Long:  `Download an image from the public or private repo`,
		Run: func(cmd *cobra.Command, args []string) {
			err := imageDownload(imageOsType, imageForceDownload)
			if err != nil {
				log.Fatal(err)
			}
			err = imageUnzip(imageDataset, imageOsType)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func imageUnzip(imageDataset string, imageOsType string) error {
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

	if !slices.Contains(hostConfig.ActiveDatasets, imageDataset) {
		return errors.New("dataset is not being used for VMs or doesn't exist")
	}

	zipFileLocation := "/tmp/" + imageOsType + ".zip"
	r, err := zip.OpenReader(zipFileLocation)
	if err != nil {
		return err
	}
	defer r.Close()

	// Iterate through the files in the archive.
	bar := progressbar.Default(
		int64(len(r.File)),
		" 📥 Unzipping the OS image || "+zipFileLocation+" || ",
	)
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer rc.Close()

		// Create the destination file.
		// dst, err := os.Create(f.Name)
		dst, err := os.Create("/tmp/disk0.img")
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer dst.Close()

		// Copy the file contents to the destination file.
		if _, err := io.Copy(dst, rc); err != nil {
			fmt.Println(err)
			continue
		}
		bar.Add(1)
	}

	return nil
}

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

		f, err := os.OpenFile("/tmp/"+osType+".zip", os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()

		bar := progressbar.DefaultBytes(
			resp.ContentLength,
			" 📥 Downloading OS image || "+vmImage+" || ",
		)
		io.Copy(io.MultiWriter(f, bar), resp.Body)
	} else {
		return errors.New("sorry, could not find the image")
	}

	return nil
}
