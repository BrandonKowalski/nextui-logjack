package ui

import (
	"errors"
	"fmt"
	"nextui-logjack/utils"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
)

type MainScreenInput struct {
	ServerURL string
}

type MainScreenOutput struct{}

type MainScreen struct{}

func NewMainScreen() *MainScreen {
	return &MainScreen{}
}

func (s *MainScreen) Draw(input MainScreenInput) (ScreenResult[MainScreenOutput], error) {
	output := MainScreenOutput{}

	sections := s.buildSections(input.ServerURL)

	options := gaba.DefaultInfoScreenOptions()
	options.Sections = sections
	options.ShowThemeBackground = false
	options.ShowScrollbar = false

	_, err := gaba.DetailScreen("", options, []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Quit"},
	})

	if err != nil {
		if errors.Is(err, gaba.ErrCancelled) {
			return quit(output), nil
		}
		gaba.GetLogger().Error("Main screen error", "error", err)
		return withAction(output, ActionError), err
	}

	return quit(output), nil
}

func (s *MainScreen) buildSections(serverURL string) []gaba.Section {
	sections := make([]gaba.Section, 0)

	qrcode, err := utils.CreateTempQRCode(serverURL, int(gaba.GetWindow().GetWidth()/3))
	sections = append(sections, gaba.NewDescriptionSection(
		"Log Jack",
		fmt.Sprintf("Scan the QR code or visit following URL to access logs from your device: %s", serverURL),
	))

	if err == nil {
		sections = append(sections, gaba.NewImageSection(
			"",
			qrcode,
			gaba.GetWindow().GetWidth()/3,
			gaba.GetWindow().GetWidth()/3,
			constants.TextAlignCenter,
		))
	} else {
		gaba.GetLogger().Error("Unable to generate QR code", "error", err)

		sections = append(sections, gaba.NewDescriptionSection(
			"",
			fmt.Sprintf("Visit the following URL to rip logs from your device: %s", serverURL),
		))
	}

	return sections
}
