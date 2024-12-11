package org

import (
	"org-service/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type OrgHTTPTransport interface {
	AddOrg(c *fiber.Ctx) error
	GetOrgs(c *fiber.Ctx) error
}

type orgHttpTransport struct {
	orgApi OrgAPI
	logger log.AllLogger
}

func NewOrgHTTPTransport(orgApi OrgAPI, logger log.AllLogger) *orgHttpTransport {
	return &orgHttpTransport{orgApi: orgApi, logger: logger}
}


func (s *orgHttpTransport) AddOrg(c *fiber.Ctx) error {
	req := &AddOrgRequest{}
	userId, err := middleware.CtxUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}

	req.UserID = userId
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	res, err := s.orgApi.AddOrg(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(res)
}

func (s *orgHttpTransport) GetOrgs(c *fiber.Ctx) error {
	req := &IDRequest{}
	userId, err := middleware.CtxUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}

	req.UserID = userId

	res, err := s.orgApi.GetOrgs(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(res)
}
