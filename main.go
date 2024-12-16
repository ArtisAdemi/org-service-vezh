package main

import (
	"os"

	_ "org-service/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"gopkg.in/gomail.v2"

	"org-service/db"
	"org-service/middleware"
	orgsvc "org-service/org"
	usersvc "org-service/users"
)

func main() {
	uiAppUrl := ""
	if os.Getenv("UI_APP_URL") != "" {
		uiAppUrl = os.Getenv("UI_APP_URL")
	}
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10 MB
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Change this to specific domains in production
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	db, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}	


	defaultLogger := log.DefaultLogger()

	authMiddleware := middleware.Authentication(os.Getenv("JWT_SECRET_KEY"))
	rbac := middleware.NewRBAC(db)

	apisRouter := app.Group("/api")
	orgRoute := apisRouter.Group("/o/:orgId", authMiddleware, rbac.OrgAccess)

	apisRouter.Get("/swagger/*", basicauth.New(basicauth.Config{
		Users: map[string]string{
			"influxo": "123123123",
		},
	}), swagger.HandlerDefault)

	// Pass gomail dialer to user service
	dialer := gomail.NewDialer("smtp.gmail.com", 587, os.Getenv("EMAIL_FROM"), os.Getenv("MAIL_PASSWORD"))


	// Initialize service
	orgApiSvc := orgsvc.NewOrgHTTPTransport(orgsvc.NewOrgService(db, defaultLogger), defaultLogger)
	userApiSvc := usersvc.NewUserHTTPTransport(usersvc.NewUserService(db, dialer, uiAppUrl))
	
	// Register routes
	orgsvc.RegisterRoutes(apisRouter, orgApiSvc, authMiddleware)
	usersvc.RegisterRoutes(apisRouter, orgRoute, userApiSvc, authMiddleware)
	
	db.AutoMigrate(
		&orgsvc.Org{},
		&orgsvc.UserOrgRole{},
	)
		app.Listen(":3002")
}