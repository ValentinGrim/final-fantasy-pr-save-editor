//go:build !windows

package io

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/storage"
	"github.com/godbus/dbus/v5"
)

const (
	portalDest     = "org.freedesktop.portal.Desktop"
	portalPath     = "/org/freedesktop/portal/desktop"
	portalOpenFile = "org.freedesktop.portal.FileChooser.OpenFile"
	portalResponse = "org.freedesktop.portal.Request.Response"
)

// picking prevents a second picker while one is open: on Wayland the portal dialog cannot be parented to our
// window, so it is not modal for it.
var picking atomic.Bool

// pickDirectory lets the user choose a directory, starting in dir, and calls onPicked (on the Fyne thread) with
// the choice; it is not called when the user cancels.
//
// The XDG desktop portal is used so that the desktop's own folder picker shows up (KDE, GNOME, ...) without
// linking GTK into the editor. When no portal is available, Fyne's built-in folder dialog is used instead.
func pickDirectory(parent fyne.Window, title, dir string, onPicked func(string)) {
	if !picking.CompareAndSwap(false, true) {
		return
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	handle := portalParentHandle(parent)
	go func() {
		picked, err := portalPickDirectory(handle, title, dir)
		fyne.Do(func() {
			defer picking.Store(false)
			if err != nil {
				fyne.LogError("desktop portal folder picker unavailable, using the built-in one", err)
				fynePickDirectory(parent, title, dir, onPicked)
			} else if picked != "" {
				onPicked(picked)
			}
		})
	}()
}

// portalParentHandle returns the portal identifier of the window on X11. Wayland would need xdg_foreign, which
// GLFW does not expose, so the dialog is left without a parent there.
func portalParentHandle(w fyne.Window) (handle string) {
	if nw, ok := w.(driver.NativeWindow); ok {
		nw.RunNative(func(ctx any) {
			if x, ok := ctx.(driver.X11WindowContext); ok && x.WindowHandle != 0 {
				handle = fmt.Sprintf("x11:%x", x.WindowHandle)
			}
		})
	}
	return
}

// portalPickDirectory calls org.freedesktop.portal.FileChooser.OpenFile in directory mode and waits for the
// answer. It returns "" with a nil error when nothing was chosen, and an error when the portal is unusable.
func portalPickDirectory(parentHandle, title, dir string) (string, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return "", err
	}
	defer func() { _ = conn.Close() }()
	if len(conn.Names()) == 0 {
		return "", errors.New("no unique name on the session bus")
	}

	// Subscribe before calling so the Response cannot be missed (see the Request interface documentation).
	token := make([]byte, 8)
	_, _ = rand.Read(token)
	handleToken := "ffpr_" + hex.EncodeToString(token)
	sender := strings.ReplaceAll(strings.TrimPrefix(conn.Names()[0], ":"), ".", "_")
	request := dbus.ObjectPath(portalPath + "/request/" + sender + "/" + handleToken)
	if err = conn.AddMatchSignal(
		dbus.WithMatchObjectPath(request),
		dbus.WithMatchInterface("org.freedesktop.portal.Request"),
		dbus.WithMatchMember("Response"),
	); err != nil {
		return "", err
	}
	signals := make(chan *dbus.Signal, 16)
	conn.Signal(signals)

	options := map[string]dbus.Variant{
		"handle_token": dbus.MakeVariant(handleToken),
		"modal":        dbus.MakeVariant(true),
		"directory":    dbus.MakeVariant(true),
		"accept_label": dbus.MakeVariant("Select"),
	}
	if dir != "" {
		options["current_folder"] = dbus.MakeVariant(append([]byte(dir), 0)) // NUL terminated byte string
	}
	var handle dbus.ObjectPath
	if err = conn.Object(portalDest, portalPath).Call(portalOpenFile, 0, parentHandle, title, options).Store(&handle); err != nil {
		return "", err
	}

	for sig := range signals {
		if sig.Name != portalResponse || (sig.Path != handle && sig.Path != request) || len(sig.Body) != 2 {
			continue
		}
		if status, _ := sig.Body[0].(uint32); status != 0 { // 1 = cancelled, 2 = closed some other way
			return "", nil
		}
		results, _ := sig.Body[1].(map[string]dbus.Variant)
		var uris []string
		if v, ok := results["uris"]; ok {
			_ = v.Store(&uris)
		}
		if len(uris) == 0 {
			return "", nil
		}
		u, err := url.Parse(uris[0])
		if err != nil || u.Scheme != "file" {
			return "", fmt.Errorf("unsupported folder URI %q", uris[0])
		}
		return u.Path, nil
	}
	return "", errors.New("session bus connection closed")
}

// fynePickDirectory is the fallback when no desktop portal is available.
func fynePickDirectory(parent fyne.Window, title, dir string, onPicked func(string)) {
	d := dialog.NewFolderOpen(func(u fyne.ListableURI, err error) {
		if err == nil && u != nil {
			onPicked(u.Path())
		}
	}, parent)
	if l, err := storage.ListerForURI(storage.NewFileURI(dir)); err == nil {
		d.SetLocation(l)
	}
	d.SetTitleText(title)
	d.SetConfirmText("Select")
	d.Resize(parent.Canvas().Size().Subtract(fyne.NewSquareSize(40)))
	d.Show()
}
