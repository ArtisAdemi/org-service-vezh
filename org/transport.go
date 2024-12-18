package org

import (
	"fmt"
	"org-service/middleware"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type OrgHTTPTransport interface {
	AddOrg(c *fiber.Ctx) error
	FindMyOrgs(c *fiber.Ctx) error
	GetOrgMembers(c *fiber.Ctx) error
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

func (s *orgHttpTransport) FindMyOrgs(c *fiber.Ctx) error {
	req := &IDRequest{}
	userId, err := middleware.CtxUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}

	req.UserID = userId

	res, err := s.orgApi.FindMyOrgs(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(res)
}

func (s *orgHttpTransport) GetOrgMembers(c *fiber.Ctx) error {
	req := &OrgRequest{}
	userId, err := middleware.CtxUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}

	orgIdStr := c.Params("orgId")
	fmt.Println("orgIdStr", orgIdStr)
	orgId, err := strconv.Atoi(orgIdStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid org id")
	}

	req.UserID = userId
	req.OrgID = orgId

	resp, err := s.orgApi.GetOrgMembers(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(resp)
}

