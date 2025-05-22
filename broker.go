// all the code for the message broker exists here
package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rivo/tview"
)

type MessageBroker struct {
	App     *tview.Application
	Rdb     *redis.Client
	Stream  string
	Handler func(sender, content string)
}

var ctx = context.Background()

func extractDetails(msg redis.XMessage) (sender string, content string, err error) {

	if contentVal, ok := msg.Values["content"].(string); ok {
		content = contentVal
	} else {
		err = fmt.Errorf("no valid `content` field present in message")
	}

	if senderVal, ok := msg.Values["sender"].(string); ok {
		sender = senderVal
	} else {
		err = fmt.Errorf("no valid `sender` field present in message")
	}
	return
}

func (messageBroker *MessageBroker) processMessage(msg redis.XMessage) {
	sender, content, err := extractDetails(msg)
	if err != nil {
		// Trace this ??
		return
	}
	messageBroker.Handler(sender, content)
}

func (messageBroker *MessageBroker) SendMessage(sender, content string) {
	_, err := messageBroker.Rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: messageBroker.Stream,
		ID:     "*",
		Values: []interface{}{"sender", sender, "content", content},
	}).Result()

	if err != nil {
		panic(err) // TODO: Handle this more gracefully
	}
}

// blocking operation - Do NOT call from the main thread
func (messageBroker *MessageBroker) StartListening() {
	id := "0"
	for {
		res, err := messageBroker.Rdb.XRead(ctx, &redis.XReadArgs{
			Block:   0,
			Streams: []string{messageBroker.Stream},
			Count:   1,
			ID:      id,
		}).Result()

		if err != nil {
			panic(err)
		}

		for i := range res {
			redisStream := res[i]
			if redisStream.Stream != messageBroker.Stream {
				continue
			}

			for i := range redisStream.Messages {
				msg := redisStream.Messages[i]
				id = msg.ID
				// in a seperate goroutine because we don't want to wait for UI update to complete
				// before continuing to process the stream - that will slow down stream processing
				go messageBroker.App.QueueUpdateDraw(func() { messageBroker.processMessage(msg) })
			}
		}
	}
}
