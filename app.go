package main

import (
	"fmt"
	"os"
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
	app             *tview.Application
	messages        []Message
	messageView     *tview.TextView
	inputField      *tview.InputField
	layout          *tview.Flex
	username        string
	currentUser     int
	currentUserName string
	messageBroker   *MessageBroker
	chatSession     *ChatSession
}

func (chatApp *ChatApp) initializeMessageBroker() {
	chatApp.messageBroker = &MessageBroker{
		App:    chatApp.app,
		Rdb:    rdb,
		Stream: "messages",
		Handler: func(sender, content string) {
			chatApp.AddMessage(sender, content)
		},
	}
}

func (chatApp *ChatApp) initializeLLM() {
	config, err := GetSystemConfig()
	if err != nil {
		panic(err)
	}
	chatApp.chatSession, err = NewChatSession(config.Model, chatApp.messageBroker)
	if err != nil {
		panic(err)
	}
}

// NewChatApp creates a new chat application instance
func NewChatApp(username string) *ChatApp {
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

	chatApp := &ChatApp{
		app:             app,
		messages:        []Message{},
		messageView:     messageView,
		inputField:      inputField,
		layout:          layout,
		username:        "You", // Joe Goldberg much??
		currentUser:     0,
		currentUserName: username,
	}
	chatApp.initializeMessageBroker()
	chatApp.initializeLLM()
	return chatApp
}

// Start initializes and runs the chat application
func (c *ChatApp) Start() error {
	// Add some initial messages
	welcomeMsg := fmt.Sprintf("Welcome to CLI-Chat %s! Type your message and press Enter to send.", c.currentUserName)
	c.AddMessage("System", welcomeMsg)

	// Handle input field events
	c.inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := c.inputField.GetText()
			if text == "" {
				return
			}

			// Add the user's message
			// c.AddMessage(Message{Sender: c.username, Content: text})
			c.messageBroker.SendMessage(c.currentUserName, text)
			c.inputField.SetText("")

			if strings.Contains(strings.ToLower(text), "..") {
				go c.chatSession.GenerateResponse(text)
			}
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
func (c *ChatApp) AddMessage(sender, content string) {
	msg := Message{
		Sender:    sender,
		Content:   content,
		Timestamp: time.Now(),
	}
	c.messages = append(c.messages, msg)

	// Format and display the message
	timestamp := msg.Timestamp.Format("15:04:05")

	var color string
	switch msg.Sender {
	case c.currentUserName:
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("You must provide your username as an argument")
		return
	}

	chat := NewChatApp(os.Args[1])
	if err := chat.Start(); err != nil {
		panic(err)
	}
}

func init() {
	// initialize the redis client & establish connection
	config, err := GetSystemConfig()
	if err != nil {
		panic(err)
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     config.RedisServerUri,
		Password: "",
		DB:       0,
	})

}
