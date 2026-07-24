package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Shared media library used by both the bound App methods and the media
	// HTTP server that streams audio/cover/lyric assets to the WebView.
	lib := NewLibrary()

	// Serve media on a dedicated loopback port. Using a separate origin (rather
	// than the Wails asset-server fallthrough) keeps streaming working in both
	// `wails dev` and production builds, since the Vite dev server would
	// otherwise answer /media requests with index.html.
	mediaBase, err := startMediaServer(lib)
	if err != nil {
		println("Error: failed to start media server:", err.Error())
	}

	app := NewApp(lib, mediaBase)

	err = wails.Run(&options.App{
		Title:     "Vinyl Player",
		Width:     1280,
		Height:    800,
		MinWidth:  980,
		MinHeight: 640,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 44, G: 33, B: 24, A: 1},
		OnStartup:        app.startup,
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
