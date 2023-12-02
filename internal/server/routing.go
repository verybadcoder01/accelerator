package server

import "github.com/gofiber/fiber/v2"

func setupRouting(app *fiber.App) {
	app.Get("/hemlo", func(c *fiber.Ctx) error {
		return c.SendString("hemlo!")
	})
	app.Get("/api/brands/open", func(c *fiber.Ctx) error { // api/brands/open?offset=x&limit=y
		return c.SendString("pizdec")
	})
}
