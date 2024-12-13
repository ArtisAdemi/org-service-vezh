package middleware

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type HTTPError struct {
	Message string `json:"message"`
}

// Role-Based Access Control
type RBAC interface {
	OrgAccess(c *fiber.Ctx) error
	RolePermissions(c *fiber.Ctx) error
}

type rbac struct {
	db *gorm.DB
}

func NewRBAC(db *gorm.DB) RBAC {
	return &rbac{
		db: db,
	}
}

func (r rbac) OrgAccess(c *fiber.Ctx) error {
	ctxUserId, err := CtxUserID(c)
	if err != nil {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(HTTPError{Message: "Invalid UserID"})
	}

	// Pull and handle orgId from URL param
	orgIdParam := c.Params("orgId")
	if orgIdParam == "" {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(HTTPError{Message: "Invalid OrgID param"})
	}

	// Check and handle in DB if relationship exists
	var userOrgRole UserOrgRole
	result := r.db.Where("user_id = ? AND org_id = ?", ctxUserId, orgIdParam).First(&userOrgRole)
	if result.Error != nil {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(HTTPError{Message: "Org access denied"})
	}

	// save the userOrgRole record ctx locals
	c.Locals("userOrgRole", userOrgRole)
	c.Next()
	return nil
}

func (r rbac) RolePermissions(c *fiber.Ctx) error {
	// Handle userOrgRole saved in ctx
	usOrgRoleI := c.Locals("userOrgRole")
	if usOrgRoleI == nil {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(HTTPError{Message: "Invalid User Role"})
	}
	usOrgRole, ok := usOrgRoleI.(UserOrgRole)
	if !ok {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(HTTPError{Message: "Error converting UserOrgRole interface to struct"})
	}

	// Collect request route's path and HTTP method
	reqRouteMethod := c.Route().Method
	reqRoutePath := c.Route().Path

	// Check in DB if the permission is allowed for usOrgRole.role_id
	var count int64
	result := r.db.Table("role_permissions").
		Where("roles.id = ? AND permissions.http_method = ? AND permissions.path = ?", usOrgRole.RoleID, reqRouteMethod, reqRoutePath).
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Count(&count)
	if result.Error != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(HTTPError{Message: result.Error.Error()})
	}
	if count == 0 {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(HTTPError{Message: "Permission denied"})
	}

	c.Next()
	return nil
}
