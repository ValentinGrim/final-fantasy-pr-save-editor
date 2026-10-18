package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	d "github.com/tawesoft/golib/v2/dialog"
	"pixel-remastered-save-editor/browser"
	"pixel-remastered-save-editor/ui"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			_ = d.Raise(fmt.Sprintf("crashed due to: %v", err))
		}
	}()
	gui := ui.New()
	gui.Load()
	var quit atomic.Bool
	go func() {
		// the update check is best effort: never let it take the app down
		defer func() { _ = recover() }()
		time.Sleep(1 * time.Second)
		if hasNewer, version, err := browser.CheckForUpdate(); err == nil && hasNewer && !quit.Load() {
			fyne.Do(func() {
				dialog.NewConfirm(
					"Update Available",
					fmt.Sprintf("A newer version of the editor is available: %s\nWould you like to download it now?", version), func(ok bool) {
						if ok {
							if e := browser.Update(version); e != nil {
								dialog.NewError(e, gui.Window()).Show()
							}
						}
					}, gui.Window()).Show()
			})
		}
	}()
	// Run returns once the driver has quit (window closed, File > Quit,
	// SIGINT/SIGTERM); GLFW is terminated by then, so don't call Quit again.
	gui.Run()
	quit.Store(true)
}
