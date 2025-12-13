package main

import (
	"context"
	"embed"
	"muu-alpha/backend/appconf"
	"muu-alpha/backend/engine"
	"muu-alpha/backend/fileloader"
	"muu-alpha/backend/pi"
	"muu-alpha/backend/system"
	"net/http"
	"os"
	"path/filepath"

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
	sysSrv := system.System()

	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exeDir := filepath.Dir(exePath)

	mux := http.NewServeMux()
	assetsDir := filepath.Join(exeDir, "static")
	assetsLoader := fileloader.New(assetsDir)
	mux.Handle("/static/", http.StripPrefix("/static/", assetsLoader))
	resDir := filepath.Join(exeDir, "resource")
	resLoader := fileloader.New(resDir)
	mux.Handle("/resource/", http.StripPrefix("/resource/", resLoader))

	err = wails.Run(&options.App{
		Title:  "muu-alpha",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: mux,
		},
		Frameless:        true,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			pi.Startup(ctx)
			appconf.Startup(ctx)
			engine.Startup(ctx)
			system.Startup(ctx)
		},
		Bind: []interface{}{
			piSrv,
			appConfSrv,
			engSrv,
			sysSrv,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
