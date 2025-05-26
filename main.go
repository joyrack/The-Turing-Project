package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joyrack/cli-chat/config"
	"github.com/joyrack/cli-chat/controllers"
	"github.com/joyrack/cli-chat/models"
	"github.com/joyrack/cli-chat/views"
	"github.com/redis/go-redis/v9"
	"github.com/rivo/tview"
)

func main() {
	// read username from the command-line args
	if len(os.Args) < 2 {
		fmt.Println("You must provide your username as an argument")
		return
	}

	var logFile string
	if len(os.Args) == 3 {
		logFile = os.Args[2]
	} else {
		logFile = "app.log"
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	logger := slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	slog.SetDefault(logger)

	config, err := config.GetSystemConfig()
	if err != nil {
		fmt.Printf("Configuration error: %v", err)
		return
	}

	app := tview.NewApplication()
	model := models.NewPlayerModel(redis.NewClient(&redis.Options{
		Addr:     config.RedisServerUri,
		Password: "",
		DB:       0,
	}))

	view := views.NewPlayerView(os.Args[1], app)
	controller := controllers.NewPlayerController(model, view)

	view.SetController(controller)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
