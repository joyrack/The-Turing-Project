package views

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/joyrack/cli-chat/models"
	"github.com/rivo/tview"
)

type welcomeScreen struct {
	welcomeDialog *tview.Modal
}

type loadingScreen struct {
	loadingWidget *tview.Modal
}

type chatScreen struct {
	messageView *tview.TextView
	inputField  *tview.InputField
	layout      *tview.Flex
}

// manages the UI components
type PlayerView struct {
	app           *tview.Application
	username      string
	controller    PlayerController
	welcomeScreen *welcomeScreen
	loadingScreen *loadingScreen
	chatScreen    *chatScreen
}

func NewPlayerView(username string, app *tview.Application) *PlayerView {
	view := &PlayerView{
		app:      app,
		username: username,
		welcomeScreen: &welcomeScreen{
			welcomeDialog: tview.NewModal(),
		},
		loadingScreen: &loadingScreen{
			loadingWidget: tview.NewModal(),
		},
		chatScreen: &chatScreen{
			messageView: tview.NewTextView(),
			inputField:  tview.NewInputField(),
			layout:      tview.NewFlex(),
		},
	}

	view.setupUI()
	return view
}

func (view *PlayerView) SetController(controller PlayerController) {
	view.controller = controller
}

func (v *PlayerView) setupUI() {
	v.welcomeScreen.welcomeDialog.
		SetText("This is the Welcome Screen").
		AddButtons([]string{models.QUESTIONER.String(), models.ANSWERER.String()}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			if buttonLabel == models.QUESTIONER.String() {
				go v.controller.InitializeGame(v.username, models.QUESTIONER)
				// v.app.SetRoot(v.chatScreen.layout, true).SetFocus(v.chatScreen.inputField)
				slog.Info("Displaying loading screen to Questioner")
				v.app.SetRoot(v.loadingScreen.loadingWidget, true)
			} else if buttonLabel == models.ANSWERER.String() {
				go v.controller.InitializeGame(v.username, models.ANSWERER)
				// v.app.SetRoot(v.chatScreen.layout, true).SetFocus(v.chatScreen.inputField)
				slog.Info("Displaying loading screen to Answerer")
				v.app.SetRoot(v.loadingScreen.loadingWidget, true)
			} else {
				// controller.cleanup()
				v.app.Stop()
			}
		})

	v.loadingScreen.loadingWidget.
		SetText("This is the loading screen")

	v.chatScreen.messageView.
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			v.app.Draw()
		}).
		SetBorder(true).
		SetTitle("Messages")

	v.chatScreen.inputField.
		SetLabel("> ").
		SetFieldWidth(0).
		SetBorder(true).
		SetTitle("Input")

	v.chatScreen.layout.
		SetDirection(tview.FlexRow).
		AddItem(v.chatScreen.messageView, 0, 1, false).
		AddItem(v.chatScreen.inputField, 3, 0, true)

	v.chatScreen.inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := v.chatScreen.inputField.GetText()
			if text == "" {
				return
			}

			go v.controller.AddMessage(&models.Message{Sender: v.username, Content: text})
			v.chatScreen.inputField.SetText("")
		}
	})

	v.app.SetRoot(v.welcomeScreen.welcomeDialog, true)
}

func (v *PlayerView) StartGame() {
	slog.Info("Starting game")
	v.app.SetRoot(v.chatScreen.layout, true).SetFocus(v.chatScreen.inputField)
	go v.controller.StartListeningForMessages()
}

func (v *PlayerView) UpdateMessageView(msg *models.Message) {
	timestamp := time.Now().Format("15:04:05")

	var color string
	switch msg.Sender {
	case v.username:
		color = "[green]"
	case "System":
		color = "[yellow]"
	default:
		color = "[blue]"
	}

	formattedMsg := fmt.Sprintf("%s %s%s[white]: %s\n",
		timestamp, color, msg.Sender, msg.Content)

	v.chatScreen.messageView.Write([]byte(formattedMsg))
	v.chatScreen.messageView.ScrollToEnd()
}

func (v *PlayerView) App() *tview.Application {
	return v.app
}
