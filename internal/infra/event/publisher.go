package event

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"

	"mikhailjbs/user-auth-service/internal/domain/user"
)

type EventPayload struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type redisPublisher struct {
	client *redis.Client
}

func NewRedisPublisher(redisURL string) user.EventPublisher {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	client := redis.NewClient(opts)
	return &redisPublisher{client: client}
}

func (p *redisPublisher) PublishUserCreated(ctx context.Context, u *user.User) error {
	payload := EventPayload{Event: "UserCreated", Data: u}
	bytes, _ := json.Marshal(payload)
	return p.client.Publish(ctx, "events.users", bytes).Err()
}

func (p *redisPublisher) PublishUserUpdated(ctx context.Context, u *user.User) error {
	payload := EventPayload{Event: "UserUpdated", Data: u}
	bytes, _ := json.Marshal(payload)
	return p.client.Publish(ctx, "events.users", bytes).Err()
}
