//go:build darwin

package main

import (
	"os"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type nativeMenuLabels struct {
	file            string
	view            string
	help            string
	newConnection   string
	openSqlite      string
	closeConnection string
	newGroup        string
	sqlHistory      string
	settings        string
	about           string
}

func currentNativeMenuLabels() nativeMenuLabels {
	if strings.HasPrefix(strings.ToLower(os.Getenv("LANG")), "en") {
		return nativeMenuLabels{
			file:            "File",
			view:            "View",
			help:            "Help",
			newConnection:   "New Connection…",
			openSqlite:      "Open SQLite File…",
			closeConnection: "Close Connection",
			newGroup:        "New Group…",
			sqlHistory:      "SQL Execution History…",
			settings:        "Settings…",
			about:           "About Data Nexus",
		}
	}
	return nativeMenuLabels{
		file:            "文件",
		view:            "视图",
		help:            "帮助",
		newConnection:   "新建连接…",
		openSqlite:      "打开 SQLite 文件…",
		closeConnection: "关闭当前连接",
		newGroup:        "新建分组…",
		sqlHistory:      "SQL 执行历史…",
		settings:        "设置…",
		about:           "关于 Data Nexus",
	}
}

func (a *App) platformApplicationMenu() *menu.Menu {
	labels := currentNativeMenuLabels()
	emit := func(event string) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, event)
		}
	}

	fileSub := menu.NewMenu()
	fileSub.AddText(labels.newConnection, keys.CmdOrCtrl("N"), func(_ *menu.CallbackData) {
		emit("app:new-connection")
	})
	fileSub.AddText(labels.openSqlite, keys.CmdOrCtrl("O"), func(_ *menu.CallbackData) {
		emit("app:open-sqlite")
	})
	fileSub.AddText(labels.closeConnection, keys.CmdOrCtrl("W"), func(_ *menu.CallbackData) {
		emit("app:close-connection")
	})
	fileSub.AddText(labels.newGroup, nil, func(_ *menu.CallbackData) {
		emit("app:new-group")
	})

	viewSub := menu.NewMenu()
	viewSub.AddText(labels.sqlHistory, nil, func(_ *menu.CallbackData) {
		emit("app:sql-history")
	})
	viewSub.AddText(labels.settings, keys.CmdOrCtrl(","), func(_ *menu.CallbackData) {
		emit("app:settings")
	})

	helpSub := menu.NewMenu()
	helpSub.AddText(labels.about, nil, func(_ *menu.CallbackData) {
		emit("app:about")
	})

	appMenu := menu.NewMenu()
	appMenu.Append(menu.AppMenu())
	appMenu.Append(menu.SubMenu(labels.file, fileSub))
	appMenu.Append(menu.EditMenu())
	appMenu.Append(menu.SubMenu(labels.view, viewSub))
	appMenu.Append(menu.WindowMenu())
	appMenu.Append(menu.SubMenu(labels.help, helpSub))
	return appMenu
}
