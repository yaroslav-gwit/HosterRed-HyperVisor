package cmd

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/spf13/cobra"
)

var (
	port int

	apiCmd = &cobra.Command{
		Use:   "api-server",
		Short: "Start an API server",
		Long:  `Start an API server on port 3000 (default).`,
		Run: func(cmd *cobra.Command, args []string) {
			StartApiServer(port)
		},
	}
)

func StartApiServer(listenPort int) {
	app := fiber.New()

	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	app.Get("/", func(fiberContext *fiber.Ctx) error {
		result := getAllVms()
		jsonResult, _ := json.Marshal(result)
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(string(jsonResult))
	})

	app.Listen("0.0.0.0:" + strconv.Itoa(listenPort))
}
