package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"org-service/db"
	orgsvc "org-service/org"
)

func main() {
	app := fiber.New()

	db, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}	
	// Migrate the schema
	db.AutoMigrate(&orgsvc.Org{})

	// Initialize service
	orgService := orgsvc.NewOrgService(db)

	// Register routes
	orgsvc.RegisterRoutes(app, orgService)

	app.Listen(":3002")
}