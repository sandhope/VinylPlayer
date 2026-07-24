package main

import (
	_ "embed"
	"runtime"

	"github.com/energye/systray"
)

// trayIconICO holds a high-contrast Windows icon tuned for small sizes on dark /
// transparent taskbars. It is intentionally separate from the executable icon
// (build/windows/icon.ico), which is optimized for larger taskbar rendering.
//
//go:embed build/windows/tray.ico
var trayIconICO []byte

// runTray creates the system-tray icon and its context menu, then blocks on the
// tray's native message loop. Because that loop relies on a per-thread Windows
// message queue, the goroutine must stay pinned to a single OS thread.
func (a *App) runTray() {
	runtime.LockOSThread()
	systray.Run(a.onTrayReady, nil)
}

// onTrayReady wires up the tray icon, tooltip and menu once systray is ready.
func (a *App) onTrayReady() {
	if len(trayIconICO) > 0 {
		systray.SetIcon(trayIconICO)
	}
	systray.SetTooltip("Vinyl Player")

	mShow := systray.AddMenuItem("显示主界面", "显示 Vinyl Player 窗口")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "退出 Vinyl Player")

	mShow.Click(func() { a.ShowWindow() })
	mQuit.Click(func() {
		// Remove the tray icon before shutting the app down.
		systray.Quit()
		a.Quit()
	})

	// Left click / double click restores the window; right click opens the menu.
	systray.SetOnClick(func(_ systray.IMenu) { a.ShowWindow() })
	systray.SetOnDClick(func(_ systray.IMenu) { a.ShowWindow() })
	systray.SetOnRClick(func(menu systray.IMenu) {
		if menu != nil {
			menu.ShowMenu()
		}
	})
}
