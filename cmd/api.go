package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/logger"
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

func StartApiServer(listenPort int, user string, password string) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true, Prefork: false})

	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

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
		vm := vmName{}
		if err := fiberContext.BodyParser(vm); err != nil {
			return err
		}
		result := getVmInfo(vm.Name)
		jsonResult, err := json.Marshal(result)
		if err != nil {
			log.Println(err)
		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(string(jsonResult))
	})

	fmt.Println("")
	fmt.Println(" Use these credentials to authenticate with the API:")
	fmt.Println(" Username:", user, "|| Password:", password)
	fmt.Println(" Address: http://0.0.0.0:" + strconv.Itoa(listenPort) + "/")
	fmt.Println("")

	app.Listen("0.0.0.0:" + strconv.Itoa(listenPort))
}
