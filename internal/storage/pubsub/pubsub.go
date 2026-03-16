package pubsub

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type EnvironmentUpdate struct {
	EnvironmentID string      `json:"environment_id"`
	Data          interface{} `json:"data,omitempty"`
}

type Publisher struct {
	rdb *redis.Client
}

func NewPublisher(rdb *redis.Client) *Publisher {
	return &Publisher{rdb: rdb}
}

func (p *Publisher) PublishEnvironmentUpdate(ctx context.Context, envID string, data interface{}) error {
	payload, err := json.Marshal(EnvironmentUpdate{
		EnvironmentID: envID,
		Data:          data,
	})
	if err != nil {
		return err
	}

	return p.rdb.Publish(ctx, "environment_updates", payload).Err()
}
