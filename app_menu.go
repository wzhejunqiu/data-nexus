//go:build !darwin

package main

import (
	"github.com/wailsapp/wails/v2/pkg/menu"
)

func (a *App) platformApplicationMenu() *menu.Menu {
	appMenu := menu.NewMenu()
	appMenu.Append(menu.AppMenu())
	appMenu.Append(menu.EditMenu())
	appMenu.Append(menu.WindowMenu())
	return appMenu
}
