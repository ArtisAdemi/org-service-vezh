package users

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(orgRouter fiber.Router, userHttpTransport UserHTTPTransport, authMiddleware func(c *fiber.Ctx) error) {
	userRouter := orgRouter.Group("/users")
	userRouter.Put("/change-user-role", userHttpTransport.ChangeUserRole)
	userRouter.Put("/change-user-status", userHttpTransport.ChangeUserStatus)
}
