package main

import (
	"context"
	"path/filepath"
	"strings"

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
	appSvc *wailssvc.AppService
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
		appSvc:    appSvc,
		mgr:       mgr,
		startupDB: startupDB,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.dialog.SetContext(ctx)
	a.appSvc.SetContext(ctx)
	runtime.OnFileDrop(ctx, a.handleFileDrop)

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

func (a *App) handleOpenDatabase() {
	if a.onOpenDatabase != nil {
		a.onOpenDatabase()
	}
}

func (a *App) handleCloseConnection() {
	id := a.appSvc.ActiveConnectionID()
	if id == "" {
		return
	}
	if err := a.conn.CloseConnection(id); err != nil {
		a.log.Warn("close connection from menu failed", zap.Error(err))
	}
	runtime.EventsEmit(a.ctx, "app:connections-changed")
}

func (a *App) handleFileDrop(_ int, _ int, paths []string) {
	for _, path := range paths {
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".db" && ext != ".sqlite" && ext != ".sqlite3" {
			continue
		}
		if _, err := a.conn.OpenConnectionFromFile(model.ConnectRequest{FilePath: path}); err != nil {
			a.log.Warn("open dropped database failed", zap.String("path", path), zap.Error(err))
			continue
		}
		runtime.EventsEmit(a.ctx, "app:connections-changed")
		return
	}
}

func (a *App) ApplicationMenu() *menu.Menu {
	fileSub := menu.NewMenu()
	fileSub.AddText("Open...", keys.CmdOrCtrl("O"), func(_ *menu.CallbackData) {
		a.handleOpenDatabase()
	})
	fileSub.AddText("Close Connection", keys.CmdOrCtrl("W"), func(_ *menu.CallbackData) {
		a.handleCloseConnection()
	})
	fileSub.AddSeparator()
	fileSub.AddText("Quit", keys.CmdOrCtrl("Q"), func(_ *menu.CallbackData) {
		runtime.Quit(a.ctx)
	})

	viewSub := menu.NewMenu()
	viewSub.AddText("Theme: Light", nil, func(_ *menu.CallbackData) {
		_ = a.appSvc.EmitThemeChange("light")
	})
	viewSub.AddText("Theme: Dark", nil, func(_ *menu.CallbackData) {
		_ = a.appSvc.EmitThemeChange("dark")
	})
	viewSub.AddText("Theme: System", nil, func(_ *menu.CallbackData) {
		_ = a.appSvc.EmitThemeChange("system")
	})
	viewSub.AddSeparator()
	viewSub.AddText("Language: 中文", nil, func(_ *menu.CallbackData) {
		_ = a.appSvc.EmitLanguageChange("zh-CN")
	})
	viewSub.AddText("Language: English", nil, func(_ *menu.CallbackData) {
		_ = a.appSvc.EmitLanguageChange("en")
	})

	helpSub := menu.NewMenu()
	helpSub.AddText("About Data Nexus", nil, func(_ *menu.CallbackData) {
		_ = a.appSvc.ShowAbout()
	})

	appMenu := menu.NewMenu()
	appMenu.Append(menu.AppMenu())
	appMenu.Append(menu.EditMenu())
	appMenu.Append(menu.WindowMenu())
	appMenu.Append(menu.SubMenu("File", fileSub))
	appMenu.Append(menu.SubMenu("View", viewSub))
	appMenu.Append(menu.SubMenu("Help", helpSub))
	return appMenu
}
