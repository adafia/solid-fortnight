package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/adafia/solid-fortnight/internal/protocol"
	"github.com/redis/go-redis/v9"
)

type RedisStreamProcessor struct {
	client     *redis.Client
	streamName string
}

func NewRedisStreamProcessor(client *redis.Client, streamName string) *RedisStreamProcessor {
	return &RedisStreamProcessor{
		client:     client,
		streamName: streamName,
	}
}

func (p *RedisStreamProcessor) Process(events []protocol.EvaluationEvent) error {
	ctx := context.Background()
	
	// Create a pipeline to send all events in a single round trip
	pipe := p.client.Pipeline()
	
	for _, event := range events {
		eventData, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
			continue // Skip bad events but keep processing others
		}

		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: p.streamName,
			Values: map[string]interface{}{
				"event": eventData,
			},
		})
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("Failed to publish events to Redis stream: %v", err)
		return err
	}

	log.Printf("Successfully queued %d events to %s", len(events), p.streamName)
	return nil
}
