package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/adafia/solid-fortnight/internal/protocol"
	"github.com/adafia/solid-fortnight/internal/storage/store"
	"github.com/redis/go-redis/v9"
)

type EvaluationEventWorker struct {
	redisClient *redis.Client
	eventStore  *store.EvaluationEventStore
	streamName  string
	groupName   string
	consumerID  string
}

func NewEvaluationEventWorker(
	redisClient *redis.Client,
	eventStore *store.EvaluationEventStore,
	streamName string,
	groupName string,
	consumerID string,
) *EvaluationEventWorker {
	return &EvaluationEventWorker{
		redisClient: redisClient,
		eventStore:  eventStore,
		streamName:  streamName,
		groupName:   groupName,
		consumerID:  consumerID,
	}
}

func (w *EvaluationEventWorker) Start(ctx context.Context) error {
	// 1. Create consumer group (if it doesn't exist)
	err := w.redisClient.XGroupCreateMkStream(ctx, w.streamName, w.groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Printf("Failed to create consumer group: %v", err)
		return err
	}

	log.Printf("Worker started, consuming from stream: %s", w.streamName)

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker shutting down...")
			return ctx.Err()
		default:
			// 2. Read events from the stream
			entries, err := w.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    w.groupName,
				Consumer: w.consumerID,
				Streams:  []string{w.streamName, ">"},
				Count:    100,
				Block:    5 * time.Second,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					continue // No new messages
				}
				log.Printf("Error reading from Redis stream: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			for _, streamEntries := range entries {
				var events []protocol.EvaluationEvent
				var messageIDs []string

				for _, xMessage := range streamEntries.Messages {
					var event protocol.EvaluationEvent
					eventJSON, ok := xMessage.Values["event"].(string)
					if !ok {
						log.Printf("Skipping message %s: missing event data", xMessage.ID)
						continue
					}

					if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
						log.Printf("Failed to unmarshal event %s: %v", xMessage.ID, err)
						continue
					}

					events = append(events, event)
					messageIDs = append(messageIDs, xMessage.ID)
				}

				if len(events) > 0 {
					// 3. Batch save events to PostgreSQL
					if err := w.eventStore.SaveEvaluationEvents(ctx, events); err != nil {
						log.Printf("Failed to save events to database: %v", err)
						continue // Do not acknowledge if database save fails
					}

					// 4. Acknowledge processed messages
					if err := w.redisClient.XAck(ctx, w.streamName, w.groupName, messageIDs...).Err(); err != nil {
						log.Printf("Failed to acknowledge messages: %v", err)
					}
				}
			}
		}
	}
}
