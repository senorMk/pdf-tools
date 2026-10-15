package main

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	desktop := newDesktop()
	err := wails.Run(&options.App{
		Title: "PDF Tools", Width: 1040, Height: 780, MinWidth: 760, MinHeight: 660,
		BackgroundColour: options.NewRGB(247, 248, 250),
		AssetServer:      &assetserver.Options{Assets: assets},
		DragAndDrop:      &options.DragAndDrop{EnableFileDrop: true, DisableWebViewDrop: true},
		Mac:              &mac.Options{TitleBar: mac.TitleBarHidden(), Appearance: mac.NSAppearanceNameAqua},
		OnStartup:        func(ctx context.Context) { desktop.ctx = ctx },
		Bind:             []interface{}{desktop},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "PDF Tools could not start.")
		os.Exit(1)
	}
}
