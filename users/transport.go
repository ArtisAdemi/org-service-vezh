package users

import (
	"org-service/middleware"

	"github.com/gofiber/fiber/v2"
)

type UserHTTPTransport interface {
	ChangeUserRole(c *fiber.Ctx) error
	ChangeUserStatus(c *fiber.Ctx) error
}

type userHTTPTransport struct {
	userApi UserAPI
}

func NewUserHTTPTransport(userApi UserAPI) UserHTTPTransport {
	return &userHTTPTransport{userApi: userApi}
}


func (s *userHTTPTransport) ChangeUserRole(c *fiber.Ctx) error {
	req := &ChangeUserRoleRequest{}
	req.OrgID = middleware.CtxOrgID(c)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	resp, err := s.userApi.ChangeUserRole(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(resp)

}

func (s *userHTTPTransport) ChangeUserStatus(c *fiber.Ctx) error {
	req := &ChangeUserStatusRequest{}
	req.OrgID = middleware.CtxOrgID(c)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	resp, err := s.userApi.ChangeUserStatus(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
