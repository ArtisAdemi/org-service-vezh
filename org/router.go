package org

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, orgRoute fiber.Router, orgHttpApi OrgHTTPTransport, authMiddleware func(c *fiber.Ctx) error) {
	orgRoutes := router.Group("/orgs")
	orgRoutes.Post("/", authMiddleware, orgHttpApi.AddOrg)
	orgRoutes.Get("/me", authMiddleware, orgHttpApi.FindMyOrgs)

	orgRoute.Get("/members", authMiddleware, orgHttpApi.GetOrgMembers)
}