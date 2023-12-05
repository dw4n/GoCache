package main

import (
	"gocache/model"
	"gocache/pkg/cacheManager"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

// Define an in-memory list of users for demonstration purposes
var Users []model.User

func main() {

	// Populate Users
	Users = []model.User{
		{
			ID:    "1",
			Name:  "Tommy Vercetti",
			Email: "tommy@example.com",
		},
		{
			ID:    "2",
			Name:  "CJ (Carl Johnson)",
			Email: "cj@example.com",
		},
		{
			ID:    "3",
			Name:  "Niko Bellic",
			Email: "niko@example.com",
		},
		// Add more characters as needed
	}

	// Load environment variables from .env
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get the Redis connection string from environment variables
	redisHost := os.Getenv("REDIS_HOST")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Initialize Redis client with the connection string
	redisClient := cacheManager.InitializeRedisClient(redisHost, redisPassword)

	// Initialize Fiber
	app := fiber.New(fiber.Config{
		AppName:           "GoCache",
		EnablePrintRoutes: true,
	})

	// Middleware to inject the Redis client into the Fiber context's Locals
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("redisClient", redisClient)
		return c.Next()
	})

	// Setup routes
	SetupRoutes(app)

	// Start the Fiber server
	log.Fatal(app.Listen(":3000"))
}
