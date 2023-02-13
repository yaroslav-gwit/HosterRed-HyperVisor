package cmd

import (
	"encoding/json"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/gofiber/fiber/v2"
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

	app.Get("/", func(c *fiber.Ctx) error {
		result := GetAllVms()
		jsonResult, _ := json.Marshal(result)
		return c.SendString(string(jsonResult))
	})

	app.Listen("0.0.0.0:" + strconv.Itoa(listenPort))
}
