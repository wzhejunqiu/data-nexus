package main

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	wailssvc "github.com/wzhejunqiu/data-nexus/internal/wails"
	"go.uber.org/zap"
)

type App struct {
	ctx    context.Context
	log    *zap.Logger
	conn   *wailssvc.ConnectionService
	dialog *wailssvc.DialogService
	export *wailssvc.ExportService
	appSvc *wailssvc.AppService
	mgr    interface {
		CloseAll()
		PersistOpenConnections() error
		RestoreConnectionsOnStartup(context.Context) error
	}
	startupDB string
}

func NewApp(
	log *zap.Logger,
	conn *wailssvc.ConnectionService,
	dialog *wailssvc.DialogService,
	export *wailssvc.ExportService,
	appSvc *wailssvc.AppService,
	mgr interface {
		CloseAll()
		PersistOpenConnections() error
		RestoreConnectionsOnStartup(context.Context) error
	},
	startupDB string,
) *App {
	return &App{
		log:       log,
		conn:      conn,
		dialog:    dialog,
		export:    export,
		appSvc:    appSvc,
		mgr:       mgr,
		startupDB: startupDB,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.dialog.SetContext(ctx)
	a.export.SetContext(ctx)
	a.appSvc.SetContext(ctx)
	runtime.OnFileDrop(ctx, a.handleFileDrop)

	if a.startupDB != "" {
		if conn, err := a.conn.OpenConnectionFromFile(model.ConnectRequest{FilePath: a.startupDB}); err != nil {
			a.log.Warn("failed to open startup database", zap.String("path", a.startupDB), zap.Error(err))
			a.emitError(err)
		} else {
			a.emitConnectionOpened(conn.ID)
		}
	} else if err := a.mgr.RestoreConnectionsOnStartup(ctx); err != nil {
		a.log.Warn("failed to restore connections", zap.Error(err))
	}
}

func (a *App) shutdown(_ context.Context) {
	if err := a.mgr.PersistOpenConnections(); err != nil {
		a.log.Warn("failed to persist open connections", zap.Error(err))
	}
	a.mgr.CloseAll()
}

func (a *App) emitConnectionOpened(connectionID string) {
	if a.ctx == nil || connectionID == "" {
		return
	}
	runtime.EventsEmit(a.ctx, "app:connection-opened", map[string]string{"id": connectionID})
	runtime.EventsEmit(a.ctx, "app:connections-changed")
}

func (a *App) emitToast(message, variant string) {
	if a.ctx == nil || message == "" {
		return
	}
	runtime.EventsEmit(a.ctx, "app:toast", map[string]string{
		"message": message,
		"variant": variant,
	})
}

func (a *App) emitError(err error) {
	if err == nil {
		return
	}
	msg := err.Error()
	if appErr, ok := err.(*model.AppError); ok && appErr.Message != "" {
		msg = appErr.Message
	}
	a.emitToast(msg, "error")
}

func (a *App) handleFileDrop(x, y int, paths []string) {
	if a.ctx == nil {
		return
	}
	dbPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".db" || ext == ".sqlite" || ext == ".sqlite3" {
			dbPaths = append(dbPaths, path)
		}
	}
	if len(dbPaths) == 0 {
		return
	}
	runtime.EventsEmit(a.ctx, "app:file-drop", map[string]interface{}{
		"x":     x,
		"y":     y,
		"paths": dbPaths,
	})
}

func (a *App) ApplicationMenu() *menu.Menu {
	return a.platformApplicationMenu()
}
