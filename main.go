package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"org-service/db"
	"org-service/middleware"
	orgsvc "org-service/org"
)

func main() {
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

	apisRouter := app.Group("/api")

	// Migrate the schema
	
	orgApiSvc := orgsvc.NewOrgHTTPTransport(orgsvc.NewOrgService(db, defaultLogger), defaultLogger)
	
	// Initialize service
	
	// Register routes
	orgsvc.RegisterRoutes(apisRouter, orgApiSvc, authMiddleware)
	
	db.AutoMigrate(&orgsvc.Org{})
	app.Listen(":3002")
}