package main

import (
	"context"
	"embed"
	"muu-alpha/backend/appconf"
	"muu-alpha/backend/engine"
	"muu-alpha/backend/pi"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	piSrv := pi.PI()
	appConfSrv := appconf.AppConf()
	engSrv := engine.Engine()

	err := wails.Run(&options.App{
		Title:  "muu-alpha",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Frameless:        true,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			pi.Startup(ctx)
			appconf.Startup(ctx)
			engine.Startup(ctx)
		},
		Bind: []interface{}{
			piSrv,
			appConfSrv,
			engSrv,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
