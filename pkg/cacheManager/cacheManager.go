package cacheManager

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client
var redisOnce sync.Once

func InitializeRedisClient(host string, password string) *redis.Client {
	redisOnce.Do(func() {
		if host == "" {
			redisClient = nil // Disabled Redis
			return
		}

		op := &redis.Options{Addr: host, Password: password, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}, WriteTimeout: 5 * time.Second}
		client := redis.NewClient(op)

		// Ping the Redis server to check the connection
		pong, err := client.Ping(context.Background()).Result()
		if err != nil {
			log.Fatalf("Failed to ping Redis: %v", err)
		}
		log.Printf("Connected to Redis: %s", pong)

		redisClient = client
	})

	return redisClient
}

func GetCache(client *redis.Client, prefix string, contentKey string) (string, error) {
	// Check if the client is nil
	if client == nil {
		return "Caching is disabled (Redis client is nil)", nil
	}

	hashedKey := SHA1Hash(contentKey)
	key := prefix + "-" + hashedKey

	ctx := context.Background()
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		log.Printf("Error getting cache: %v", err)
		return "", err
	}
	return val, nil
}

func SetCache(client *redis.Client, prefix string, contentKey string, value string, expiration time.Duration) error {
	// Check if the client is nil
	if client == nil {
		log.Println("Warning: Caching is disabled (Redis client is nil)")
		return nil
	}

	hashedKey := SHA1Hash(contentKey)
	key := prefix + "-" + hashedKey

	ctx := context.Background()
	err := client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		log.Printf("Error setting cache: %v", err)
		return err
	}
	return nil
}

func RemoveCache(client *redis.Client, prefix string, contentKey string) error {
	// Check if the client is nil
	if client == nil {
		log.Println("Warning: Caching is disabled (Redis client is nil)")
		return nil
	}

	hashedKey := SHA1Hash(contentKey)
	key := prefix + "-" + hashedKey

	ctx := context.Background()
	err := client.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}

func SHA1Hash(input string) string {
	// Create a new SHA1 hash instance
	sha1Hash := sha1.New()

	// Write the input string to the hash
	sha1Hash.Write([]byte(input))

	// Get the final hash as a byte slice
	hashBytes := sha1Hash.Sum(nil)

	// Convert the byte slice to a hexadecimal string
	hashString := hex.EncodeToString(hashBytes)

	return hashString
}
