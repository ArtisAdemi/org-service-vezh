package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func Authentication(secretKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).SendString("Missing or invalid token")
		}

		// Remove "Bearer " prefix if present
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid token")
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Locals("userID", claims["userId"])
			return c.Next()
		} else {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid token")
		}
	}
}

func CtxUserID(c *fiber.Ctx) (int, error) {
	userID, ok := c.Locals("userID").(float64) // JWT claims are often float64
	if !ok {
		return 0, errors.New("user ID not found in context")
	}
	return int(userID), nil
}

func CtxOrgID(c *fiber.Ctx) int {
	usOrgRoleI := c.Locals("userOrgRole")
	if usOrgRoleI == nil {
		return 0
	}
	usOrgRole, ok := usOrgRoleI.(UserOrgRole)
	if !ok {
		return 0
	}
	return usOrgRole.OrgID
}

func CtxRoleID(c *fiber.Ctx) int {
	usOrgRoleI := c.Locals("userOrgRole")
	if usOrgRoleI == nil {
		return 0
	}
	usOrgRole, ok := usOrgRoleI.(UserOrgRole)
	if !ok {
		return 0
	}
	return usOrgRole.RoleID
}