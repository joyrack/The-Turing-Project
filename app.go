package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rivo/tview"
)

var rdb *redis.Client

// Message represents a chat message
type Message struct {
	Sender    string
	Content   string
	Timestamp time.Time
}

// ChatApp manages the chat application state
type ChatApp struct {
	app            *tview.Application
	messages       []Message
	messageView    *tview.TextView
	inputField     *tview.InputField
	layout         *tview.Flex
	username       string
	otherUsernames []string
	currentUser    int
	messageBroker  *MessageBroker
}

func (chatApp *ChatApp) initializeMessageBroker() {
	chatApp.messageBroker = &MessageBroker{
		App:    chatApp.app,
		Rdb:    rdb,
		Stream: "messages",
		Handler: func(message Message) {
			chatApp.AddMessage(message)
		},
	}
}

// NewChatApp creates a new chat application instance
func NewChatApp() *ChatApp {
	app := tview.NewApplication()

	messageView := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	messageView.SetBorder(true).SetTitle("Messages")
	messageView.SetScrollable(true)

	inputField := tview.NewInputField().
		SetLabel("> ").
		SetFieldWidth(0)
	inputField.SetBorder(true).SetTitle("Input")

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(messageView, 0, 1, false).
		AddItem(inputField, 3, 0, true)

	// Sample users
	otherUsernames := []string{"Alice", "Bob", "Charlie"}

	chatApp := &ChatApp{
		app:            app,
		messages:       []Message{},
		messageView:    messageView,
		inputField:     inputField,
		layout:         layout,
		username:       "You",
		otherUsernames: otherUsernames,
		currentUser:    0,
	}
	chatApp.initializeMessageBroker()
	return chatApp
}

// Start initializes and runs the chat application
func (c *ChatApp) Start() error {
	// Add some initial messages
	c.AddMessage(Message{Sender: "System", Content: "Welcome to Go Chat! Type your message and press Enter to send."})
	c.AddMessage(Message{Sender: "Alice", Content: "Hello there!"})
	c.AddMessage(Message{Sender: "Bob", Content: "Hey everyone!"})

	// Handle input field events
	c.inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := c.inputField.GetText()
			if text == "" {
				return
			}

			// Add the user's message
			c.AddMessage(Message{Sender: c.username, Content: text})
			c.inputField.SetText("")

			// Simulate a response from another user after a delay
			go func() {
				time.Sleep(1 * time.Second)
				otherUser := c.otherUsernames[c.currentUser]
				c.currentUser = (c.currentUser + 1) % len(c.otherUsernames)
				c.AddMessage(Message{Sender: otherUser, Content: c.generateResponse(text)})
			}()
		}
	})

	// Set up key bindings
	c.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			c.app.Stop()
			return nil
		}
		return event
	})

	// Start the message broker on a seperate goroutine
	go c.messageBroker.StartListening()

	// Start the application
	return c.app.SetRoot(c.layout, true).Run()
}

// AddMessage adds a new message to the chat history
func (c *ChatApp) AddMessage(msg Message) {
	// msg := Message{
	// 	Sender:    sender,
	// 	Content:   content,
	// 	Timestamp: time.Now(),
	// }
	msg.Timestamp = time.Now()
	c.messages = append(c.messages, msg)

	// Format and display the message
	timestamp := msg.Timestamp.Format("15:04:05")

	var color string
	switch msg.Sender {
	case "You":
		color = "[green]"
	case "System":
		color = "[yellow]"
	default:
		color = "[blue]"
	}

	formattedMsg := fmt.Sprintf("%s %s%s[white]: %s\n",
		timestamp, color, msg.Sender, msg.Content)

	c.messageView.Write([]byte(formattedMsg))

	// Auto-scroll to the bottom
	c.messageView.ScrollToEnd()
}

// generateResponse creates a simple response to user input
func (c *ChatApp) generateResponse(input string) string {
	input = strings.ToLower(input)

	if strings.Contains(input, "hello") || strings.Contains(input, "hi") {
		return "Hello there! How are you?"
	} else if strings.Contains(input, "how are you") {
		return "I'm doing well, thanks for asking!"
	} else if strings.Contains(input, "bye") || strings.Contains(input, "goodbye") {
		return "Goodbye! Talk to you later."
	} else if strings.Contains(input, "?") {
		return "That's an interesting question. Let me think about it."
	} else {
		responses := []string{
			"I understand.",
			"Interesting point.",
			"Thanks for sharing that.",
			"Let's discuss that further.",
			"I see what you mean.",
		}
		return responses[time.Now().UnixNano()%int64(len(responses))]
	}
}

func main() {
	chat := NewChatApp()
	if err := chat.Start(); err != nil {
		panic(err)
	}
}

func init() {
	// initialize the redis client & establish connection
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}
