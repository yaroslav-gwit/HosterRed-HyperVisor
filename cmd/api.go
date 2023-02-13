package cmd

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/spf13/cobra"
)

var (
	apiCmd = &cobra.Command{
		Use:   "api-server",
		Short: "Start an API server",
		Long:  `Start an API server on port 3000 (default).`,
		Run: func(cmd *cobra.Command, args []string) {
			StartApiServer(3000)
		},
	}
)

func StartApiServer(listenPort int) {
	app := fiber.New()
	// app.Use(logger.New())
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))
	app.Get("/", func(c *fiber.Ctx) error {
		result := GetAllVms()
		jsonResult, _ := json.Marshal(result)
		return c.SendString(string(jsonResult))
	})

	app.Listen("0.0.0.0:" + strconv.Itoa(listenPort))
}
