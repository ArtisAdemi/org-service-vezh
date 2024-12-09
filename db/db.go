package db

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := os.Getenv(name)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func ConnectDB() (*gorm.DB, error) {
	// Load .env file only in development
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			fmt.Print("Error loading .env file")
		}
	}

	var dbhost, dbuser, dbpassword, dbname string
	var dbport int

	// Determine which database configuration to use based on the ENV variable
	fmt.Println("Env", os.Getenv("ENV"))
	var dsn string
	switch os.Getenv("ENV") {
	case "test":
		dbhost = os.Getenv("TEST_DB_HOST")
		dbport = getEnvAsInt("TEST_DB_PORT", 5432)
		dbuser = os.Getenv("TEST_DB_USERNAME")
		dbpassword = os.Getenv("TEST_DB_PASSWORD")
		dbname = os.Getenv("TEST_DB_NAME")
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d  TimeZone=Asia/Jakarta", dbhost, dbuser, dbpassword, dbname, dbport)
	default:
		dbhost = os.Getenv("DB_HOST")
		dbport = getEnvAsInt("DB_PORT", 5432)
		dbuser = os.Getenv("DB_USERNAME")
		dbpassword = os.Getenv("DB_PASSWORD")
		dbname = os.Getenv("DB_NAME")
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=require TimeZone=Asia/Jakarta", dbhost, dbuser, dbpassword, dbname, dbport)
	}

	fmt.Printf("Connecting to DB at %s:%d with user %s\n", dbhost, dbport, dbuser)

	// Use pgx as the driver

	// Open the database connection using postgres with pgx
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Enable detailed logging
	})
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return nil, err
	}

	// Set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
