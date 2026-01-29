package redis

import (
	"context"
	"encoding/json"
	"log"
	"messaging-service/internal/config"

	"github.com/go-redis/redis/v8"
)

var Client *redis.Client
var Ctx = context.Background()

func Init(cfg *config.Config) {
	Client = redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
}

type OnlineUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func MarkOnline(userID string, userData OnlineUser) error {
	// Store user data as JSON in a hash
	userDataJSON, err := json.Marshal(userData)
	if err != nil {
		return err
	}

	// Store in hash: online_users_data:userID -> user JSON
	err = Client.HSet(Ctx, "online_users_data", userID, userDataJSON).Err()
	if err != nil {
		return err
	}

	// Also keep the simple set for quick lookups
	return Client.SAdd(Ctx, "online_users", userID).Err()
}

func MarkOffline(userID string) error {
	// Remove from both hash and set
	Client.HDel(Ctx, "online_users_data", userID)
	return Client.SRem(Ctx, "online_users", userID).Err()
}

func IsOnline(userID string) bool {
	res, err := Client.SIsMember(Ctx, "online_users", userID).Result()
	if err != nil {
		log.Println("Redis IsOnline error:", err)
		return false
	}
	return res
}

func GetOnlineUsers() ([]OnlineUser, error) {
	// Get all user data from hash
	userDataMap, err := Client.HGetAll(Ctx, "online_users_data").Result()
	if err != nil {
		return nil, err
	}

	var onlineUsers []OnlineUser
	for _, userDataJSON := range userDataMap {
		var user OnlineUser
		if err := json.Unmarshal([]byte(userDataJSON), &user); err == nil {
			onlineUsers = append(onlineUsers, user)
		}
	}

	return onlineUsers, nil
}
