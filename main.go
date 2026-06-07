package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
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

	devMode := true
	log, err := logger.New(cfg.Log, devMode)
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

	app := NewApp(log, connWails, dialogWails, mgr, cli.DBPath)
	app.SetOpenDatabaseHandler(func() {
		path, err := dialogWails.OpenDatabaseFile()
		if err != nil {
			if appErr, ok := err.(*model.AppError); ok && appErr.Code == "DIALOG_CANCELLED" {
				return
			}
			log.Warn("open database dialog failed", zap.Error(err))
			return
		}
		if _, err := connWails.OpenConnectionFromFile(model.ConnectRequest{FilePath: path}); err != nil {
			log.Warn("open connection failed", zap.Error(err))
		}
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
		Mac: &mac.Options{
			TitleBar: mac.TitleBarDefault(),
		},
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
