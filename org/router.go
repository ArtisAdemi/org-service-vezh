package org

import (
	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, service *OrgService) {
	app.Post("/orgs", func(c *fiber.Ctx) error {
		name := c.FormValue("name")
		size := c.FormValue("size")
		org, err := service.AddOrg(name, size)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(org)
	})
}