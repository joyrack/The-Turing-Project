// all the code for the message broker exists here
package main

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/rivo/tview"
)

type MessageBroker struct {
	App     *tview.Application
	Rdb     *redis.Client
	Stream  string
	Handler func(message Message)
}

var ctx = context.Background()

// Helper function to convert redis.XMessage to application Message object
func convertToMessage(msg redis.XMessage) Message {
	var message Message

	if content, ok := msg.Values["content"].(string); ok {
		message.Content = content
	} else {
		panic("Valid `content` attribute not found")
	}

	if sender, ok := msg.Values["sender"].(string); ok {
		message.Sender = sender
	} else {
		panic("Valid `sender` attribute not found")
	}

	return message
}

func (messageBroker *MessageBroker) processMessage(msg redis.XMessage) {
	message := convertToMessage(msg)
	messageBroker.Handler(message)
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
