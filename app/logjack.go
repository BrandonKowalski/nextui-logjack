package main

import (
	"nextui-logjack/internal"
	"nextui-logjack/server"
	"nextui-logjack/ui"
	"nextui-logjack/utils"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/router"
)

const (
	screenMain router.Screen = iota
)

func init() {
	gaba.Init(gaba.Options{
		WindowTitle:    "Log Jack",
		ShowBackground: true,
		LogFilename:    "logjack.log",
		IsNextUI:       true,
	})
}

func main() {
	defer gaba.Close()

	logger := gaba.GetLogger()
	logger.Info("Starting Log Jack")

	localIP := utils.GetLocalIP()
	if localIP == "127.0.0.1" {
		logger.Error("No network connection")
		gaba.ConfirmationMessage("Please connect to WiFi and try again.", []gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: "Quit"},
		}, gaba.MessageOptions{})
		return
	}

	config, err := internal.LoadConfig()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
	}

	srv := server.New(config)
	if err := srv.Start(); err != nil {
		logger.Error("Failed to start server", "error", err)
		gaba.ConfirmationMessage("Failed to start web server!", []gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: "Quit"},
		}, gaba.MessageOptions{})
		return
	}
	defer srv.Stop()

	logger.Info("Server started", "url", srv.URL())

	r := buildRouter(srv)
	if err := r.Run(screenMain, nil); err != nil {
		logger.Error("Router error", "error", err)
	}
}

func buildRouter(srv *server.Server) *router.Router {
	r := router.New()

	r.Register(screenMain, func(input any) (any, error) {
		screen := ui.NewMainScreen()
		result, err := screen.Draw(ui.MainScreenInput{
			ServerURL: srv.URL(),
		})
		return result, err
	})

	r.OnTransition(func(from router.Screen, result any, stack *router.Stack) (router.Screen, any) {
		return router.ScreenExit, nil
	})

	return r
}
