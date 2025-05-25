package views

import (
	"github.com/gdamore/tcell/v2"
	"github.com/joyrack/cli-chat/models"
	"github.com/rivo/tview"
)

type welcomeScreen struct {
	welcomeDialog *tview.Modal
}

type loadingScreen struct {
	loadingWidget *tview.TextView
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
			loadingWidget: tview.NewTextView(),
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
				v.controller.InitializeGame(v.username, models.QUESTIONER)
				// v.app.SetRoot(v.chatScreen.layout, true).SetFocus(v.chatScreen.inputField)
				v.app.SetRoot(v.loadingScreen.loadingWidget, true)
			} else if buttonLabel == models.ANSWERER.String() {
				v.controller.InitializeGame(v.username, models.ANSWERER)
				// v.app.SetRoot(v.chatScreen.layout, true).SetFocus(v.chatScreen.inputField)
				v.app.SetRoot(v.loadingScreen.loadingWidget, true)
			} else {
				// controller.cleanup()
				v.app.Stop()
			}
		})

	v.loadingScreen.loadingWidget.
		SetTitle("Loading...")

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

	v.chatScreen.inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := v.chatScreen.inputField.GetText()
			if text == "" {
				return
			}

			v.controller.AddMessage(&models.Message{Sender: v.username, Content: text})
			v.chatScreen.inputField.SetText("")
		}
	})
}
