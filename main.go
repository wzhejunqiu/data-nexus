package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wzhejunqiu/data-nexus/internal/config"
	"github.com/wzhejunqiu/data-nexus/internal/executionlog"
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

	logMgr, err := logger.NewManager(cfg.Log, config.IsDevMode())
	if err != nil {
		panic(err)
	}
	log := logMgr.Logger()
	defer func() { _ = log.Sync() }()

	store, err := service.NewConnectionStore("")
	if err != nil {
		log.Fatal("connection store", zap.Error(err))
	}
	queryStore, err := service.NewQueryStore("")
	if err != nil {
		log.Fatal("query store", zap.Error(err))
	}
	mgr := service.NewConnectionManager(store, log)
	execLog, err := executionlog.NewStore(config.DefaultExecutionLogConfig())
	if err != nil {
		log.Fatal("execution log store", zap.Error(err))
	}
	defer func() { _ = execLog.Close() }()
	querySvc := service.NewQueryService(mgr, execLog, log)
	sqlExecutionSvc := service.NewSqlExecutionService(execLog)
	cannedQuerySvc := service.NewCannedQueryService(queryStore)
	exportSvc := service.NewExportService(querySvc)
	importSvc := service.NewImportService(querySvc)

	connWails := wailssvc.NewConnectionService(mgr, log)
	schemaWails := wailssvc.NewSchemaService(querySvc, log)
	tableWails := wailssvc.NewTableService(querySvc, log)
	queryWails := wailssvc.NewQueryService(querySvc, log)
	dialogWails := wailssvc.NewDialogService(log)
	fileWails := wailssvc.NewFileService(log)
	configWails := wailssvc.NewConfigService(logMgr, log)
	exportWails := wailssvc.NewExportService(exportSvc, dialogWails, log)
	importWails := wailssvc.NewImportService(importSvc, log)
	cannedQueryWails := wailssvc.NewCannedQueryService(cannedQuerySvc, log)
	sqlExecutionWails := wailssvc.NewSqlExecutionService(sqlExecutionSvc, log)
	appWails := wailssvc.NewAppService(log)

	app := NewApp(log, connWails, dialogWails, exportWails, appWails, mgr, cli.DBPath)
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
		if conn, err := connWails.OpenConnectionFromFile(model.ConnectRequest{FilePath: path}); err != nil {
			app.emitError(err)
			log.Warn("open connection failed", zap.Error(err))
			return
		} else {
			app.emitConnectionOpened(conn.ID)
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
			fileWails,
			configWails,
			exportWails,
			importWails,
			cannedQueryWails,
			sqlExecutionWails,
			appWails,
		},
	})
	if err != nil {
		log.Fatal("wails run", zap.Error(err))
	}
}
