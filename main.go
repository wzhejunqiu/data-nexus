package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wzhejunqiu/data-nexus/internal/config"
	"github.com/wzhejunqiu/data-nexus/internal/logger"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	wailssvc "github.com/wzhejunqiu/data-nexus/internal/wails"
	"go.uber.org/zap"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	cli := config.ParseFlags()
	cfg := config.ApplyCLI(config.Load(), cli)

	log, err := logger.New(cfg.Log, config.IsDevMode())
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	store, err := service.NewConnectionStore("")
	if err != nil {
		log.Fatal("connection store", zap.Error(err))
	}
	mgr := service.NewConnectionManager(store, log)
	querySvc := service.NewQueryService(mgr)

	connWails := wailssvc.NewConnectionService(mgr, log)
	schemaWails := wailssvc.NewSchemaService(querySvc, log)
	tableWails := wailssvc.NewTableService(querySvc, log)
	queryWails := wailssvc.NewQueryService(querySvc, log)
	dialogWails := wailssvc.NewDialogService(log)
	appWails := wailssvc.NewAppService(log)

	app := NewApp(log, connWails, dialogWails, appWails, mgr, cli.DBPath)
	app.SetOpenDatabaseHandler(func() {
		path, err := dialogWails.OpenDatabaseFile()
		if err != nil {
			if appErr, ok := err.(*model.AppError); ok && appErr.Code == "DIALOG_CANCELLED" {
				return
			}
			app.emitError(err)
			log.Warn("open database dialog failed", zap.Error(err))
			return
		}
		if _, err := connWails.OpenConnectionFromFile(model.ConnectRequest{FilePath: path}); err != nil {
			app.emitError(err)
			log.Warn("open connection failed", zap.Error(err))
			return
		}
		runtime.EventsEmit(app.ctx, "app:connections-changed")
	})

	err = wails.Run(&options.App{
		Title:     "Data Nexus",
		Width:     1280,
		Height:    800,
		MinWidth:  960,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 42, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Menu:             app.ApplicationMenu(),
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarDefault(),
		},
		Linux: &linux.Options{
			ProgramName: "Data Nexus",
		},
		Windows: &windows.Options{},
		Bind: []interface{}{
			connWails,
			schemaWails,
			tableWails,
			queryWails,
			dialogWails,
			appWails,
		},
	})
	if err != nil {
		log.Fatal("wails run", zap.Error(err))
	}
}
