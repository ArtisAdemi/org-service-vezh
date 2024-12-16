package users

import (
	"net/url"
	"org-service/middleware"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type UserHTTPTransport interface {
	ChangeUserRole(c *fiber.Ctx) error
	ChangeUserStatus(c *fiber.Ctx) error
	InviteUser(c *fiber.Ctx) error
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


func (s *userHTTPTransport) InviteUser(c *fiber.Ctx) error {
	req := &InviteUserRequest{}
	req.OrgID = middleware.CtxOrgID(c)
	req.Email = c.Params("email")
	decodedEmail, err := url.QueryUnescape(req.Email)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	req.Email = decodedEmail
	roleIdParam := c.Params("roleId")
	roleId, err := strconv.Atoi(roleIdParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	req.RoleID = roleId
	userId, err := middleware.CtxUserID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	req.CurrentUserID = userId
	req.CurrentRoleID = middleware.CtxRoleID(c)

	resp, err := s.userApi.InviteUser(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
