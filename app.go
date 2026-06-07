package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
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
	mgr    interface {
		CloseAll()
		PersistOpenConnections() error
		RestoreConnectionsOnStartup(context.Context) error
	}
	startupDB      string
	onOpenDatabase func()
}

func NewApp(
	log *zap.Logger,
	conn *wailssvc.ConnectionService,
	dialog *wailssvc.DialogService,
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
		mgr:       mgr,
		startupDB: startupDB,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.dialog.SetContext(ctx)

	if a.startupDB != "" {
		if _, err := a.conn.OpenConnectionFromFile(model.ConnectRequest{FilePath: a.startupDB}); err != nil {
			a.log.Warn("failed to open startup database", zap.String("path", a.startupDB), zap.Error(err))
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

func (a *App) SetOpenDatabaseHandler(fn func()) {
	a.onOpenDatabase = fn
}

func (a *App) handleOpenDatabase() {
	if a.onOpenDatabase != nil {
		a.onOpenDatabase()
	}
}

func (a *App) ApplicationMenu() *menu.Menu {
	fileSub := menu.NewMenu()
	fileSub.AddText("Open...", keys.CmdOrCtrl("O"), func(_ *menu.CallbackData) {
		a.handleOpenDatabase()
	})
	fileSub.AddSeparator()
	fileSub.AddText("Quit", keys.CmdOrCtrl("Q"), func(_ *menu.CallbackData) {
		runtime.Quit(a.ctx)
	})

	appMenu := menu.NewMenu()
	appMenu.Append(menu.AppMenu())
	appMenu.Append(menu.EditMenu())
	appMenu.Append(menu.WindowMenu())
	appMenu.Append(menu.SubMenu("File", fileSub))
	return appMenu
}
