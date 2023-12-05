package main

import (
	"encoding/json"
	"fmt"
	"gocache/model"
	"log"
	"time"

	"gocache/pkg/cacheManager"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func SetupRoutes(app *fiber.App) {
	// GET /user
	app.Get("/usersWithoutCache", GetUserWithoutCache)

	// POST /user
	app.Post("/user", CreateUser)

	// Delete /user
	app.Delete("/user/:id", DeleteUser)

	// Delete /user
	app.Delete("/userWithoutRemovingCache/:id", DeleteUserWithoutRemovingCache)

	// GET /users
	app.Get("/users", GetUsers)
}

func GetUserWithoutCache(c *fiber.Ctx) error {
	return c.JSON(Users)
}

func DeleteUserWithoutRemovingCache(c *fiber.Ctx) error {
	// Get the user ID from the request parameters
	userID := c.Params("id")

	fmt.Println(userID)

	// Find and remove the user from the global Users slice
	for i, user := range Users {
		if user.ID == userID {
			// Remove the user from the slice
			Users = append(Users[:i], Users[i+1:]...)

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message": "User deleted successfully",
			})
		}
	}

	// User not found
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "User not found",
	})
}

func DeleteUser(c *fiber.Ctx) error {
	// Get the user ID from the request parameters
	userID := c.Params("id")

	// Find and remove the user from the global Users slice
	for i, user := range Users {
		if user.ID == userID {
			// Remove the user from the slice
			Users = append(Users[:i], Users[i+1:]...)

			// Remove the cache for the same key and prefix
			// Access the Redis client from the context's Locals
			redisClient := c.Locals("redisClient").(*redis.Client)
			//Assume prefix and content is actually the cache needed to be destroy,in real life, this is rare case
			prefix := "getusers"
			contentKey := "someUniqueIdentifiersLikeSQLQUERY"
			cacheManager.RemoveCache(redisClient, prefix, contentKey)

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message": "User deleted successfully",
			})
		}
	}

	// User not found
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "User not found",
	})
}

func CreateUser(c *fiber.Ctx) error {
	var newUser model.User

	// Parse the request body into the newUser struct
	if err := c.BodyParser(&newUser); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	newUser.ID = uuid.New().String()
	// Append the new user to the global Users slice
	Users = append(Users, newUser)

	// Remove the cache for the same key and prefix
	// Access the Redis client from the context's Locals
	redisClient := c.Locals("redisClient").(*redis.Client)
	//Assume prefix and content is actually the cache needed to be destroy,in real life, this is rare case
	prefix := "getusers"
	contentKey := "someUniqueIdentifiersLikeSQLQUERY"
	cacheManager.RemoveCache(redisClient, prefix, contentKey)

	return c.Status(fiber.StatusCreated).JSON(newUser)
}

func GetUsers(c *fiber.Ctx) error {

	// Prefix and Usually some way to set key
	prefix := "getusers"
	key := "someUniqueIdentifiersLikeSQLQUERY"

	// Access the Redis client from the context's Locals
	redisClient := c.Locals("redisClient").(*redis.Client)

	// Attempt to retrieve users from the cache
	cachedUsers, err := cacheManager.GetCache(redisClient, prefix, key)
	if err == nil && redisClient != nil {
		// Cached data found, return it
		var users []model.User
		if err := json.Unmarshal([]byte(cachedUsers), &users); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to unmarshal cached data",
			})
		}
		return c.JSON(users)
	}

	// No cached data found, fetch from "global users" data
	// For demonstration purposes, use the predefined "users" slice
	// In a real application, you would fetch data from your data source (e.g., database)
	users := Users

	// Cache the fetched data for future requests
	usersJSON, _ := json.Marshal(users)
	if err := cacheManager.SetCache(redisClient, prefix, key, string(usersJSON), 10*time.Minute); err != nil {
		// Handle the error here (e.g., log it)
		log.Printf("Error setting cache: %v", err)
		// You can choose to return an error response to the client or take other actions as needed.
	}

	return c.JSON(users)
}
