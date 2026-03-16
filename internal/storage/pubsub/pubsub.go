package pubsub

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type EnvironmentUpdate struct {
	EnvironmentID string `json:"environment_id"`
}

type Publisher struct {
	rdb *redis.Client
}

func NewPublisher(rdb *redis.Client) *Publisher {
	return &Publisher{rdb: rdb}
}

func (p *Publisher) PublishEnvironmentUpdate(ctx context.Context, envID string) error {
	payload, err := json.Marshal(EnvironmentUpdate{EnvironmentID: envID})
	if err != nil {
		return err
	}

	return p.rdb.Publish(ctx, "environment_updates", payload).Err()
}
