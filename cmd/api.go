package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/spf13/cobra"
)

var (
	apiServerPort     int
	apiServerUser     string
	apiServerPassword string

	apiCmd = &cobra.Command{
		Use:   "api-server",
		Short: "Start an API server",
		Long:  `Start an API server on port 3000 (default).`,
		Run: func(cmd *cobra.Command, args []string) {
			StartApiServer(apiServerPort, apiServerUser, apiServerPassword)
		},
	}
)

func StartApiServer(port int, user string, password string) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true, Prefork: false})
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "[${locals:requestid} - ${ip}]:${port} ${status} - ${method} ${path} - Error: ${error}\n"}))

	app.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			user: password,
		},
	}))

	app.Get("/host/info", func(fiberContext *fiber.Ctx) error {
		result := jsonOutputHostInfo()
		jsonResult, _ := json.Marshal(result)
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(string(jsonResult))
	})

	app.Get("/vm/list", func(fiberContext *fiber.Ctx) error {
		result := getAllVms()
		jsonResult, _ := json.Marshal(result)
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(string(jsonResult))
	})

	type vmName struct {
		Name string `json:"name" xml:"name" form:"name"`
	}

	app.Post("/vm/info", func(fiberContext *fiber.Ctx) error {
		vm := new(vmName)
		if err := fiberContext.BodyParser(vm); err != nil {
			return err
		}
		result, err := getVmInfo(vm.Name)
		if err != nil {
			fiberContext.Status(fiber.StatusBadRequest)
			return fiberContext.SendString(`{ "message": "` + err.Error() + `" }`)
		}
		jsonResult, err := json.Marshal(result)
		if err != nil {
			log.Println(err)
		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(string(jsonResult))
	})

	app.Post("/vm/start", func(fiberContext *fiber.Ctx) error {
		vm := new(vmName)
		if err := fiberContext.BodyParser(vm); err != nil {
			return err
		}
		err := vmStart(vm.Name)
		if err != nil {
			fiberContext.Status(fiber.StatusBadRequest)
			return fiberContext.SendString(`{ "message": "` + err.Error() + `" }`)
		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(`{ "message": "success" }`)
	})

	app.Post("/vm/start-all", func(fiberContext *fiber.Ctx) error {
		// Using NOHUP option in order to avoid killing all VMs when API server stops
		execPath, err := os.Executable()
		if err != nil {
			return fiberContext.SendString(`{ "message": "failed to start the process"}`)
		}
		execFile := path.Dir(execPath) + "/hoster"
		// Execute start all from the terminal using nohup
		cmd := exec.Command("nohup", execFile, "vm", "start-all", "&")
		err = cmd.Start()
		if err != nil {
			return fiberContext.SendString(`{ "message": "failed to start the process"}`)
		}
		go func() {
			err := cmd.Wait()
			if err != nil {
				log.Println(err)
			}
		}()

		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(`{ "message": "process started" }`)
	})

	app.Post("/vm/stop", func(fiberContext *fiber.Ctx) error {
		vm := new(vmName)
		if err := fiberContext.BodyParser(vm); err != nil {
			return err
		}
		err := vmStop(vm.Name)
		if err != nil {
			fiberContext.Status(fiber.StatusBadRequest)
			return fiberContext.SendString(`{ "message": "` + err.Error() + `" }`)
		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(`{ "message": "success" }`)
	})

	app.Post("/vm/stop-all", func(fiberContext *fiber.Ctx) error {
		go vmStopAll()
		// log.Printf("vmStopAll finished running")
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(`{ "message": "process started" }`)
	})

	fmt.Println("")
	fmt.Println(" Use these credentials to authenticate with the API:")
	fmt.Println(" Username:", user, "|| Password:", password)
	fmt.Println(" Address: http://0.0.0.0:" + strconv.Itoa(port) + "/")
	fmt.Println("")

	app.Listen("0.0.0.0:" + strconv.Itoa(port))
}
