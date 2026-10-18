# FFPR Save Editor - build for Linux and Windows.
#
# On Linux:   make build-linux | build-windows (cross, needs mingw-w64) | all
#             make package-linux | package-windows | install-linux [PREFIX=~/.local]
# On Windows: make (from an MSYS2/Git-Bash shell, with Go and a MinGW gcc on PATH)
#
# Useful overrides:
#   TAGS=...            extra Go build tags
#   UPX=                disable upx compression (it is skipped anyway if upx is missing)
#   WINDOWS_CC=...      C compiler used for the Windows build
#   PREFIX=/usr/local   install prefix for install-linux (DESTDIR is honoured too)
#   VERSION=1.4.0       version embedded in the binaries (default: latest git tag
#                       without its leading "v"; empty keeps the in-code default)
#   DIST=/some/dir      output directory (command line only)

APP      := FFPR_Save_Editor
APP_ID   := io.github.valentingrim.ffpr-save-editor
DIST     := dist

GO       ?= go
UPX      ?= upx
TAGS     ?=

# Version shown by the app and used by the update checker. Computed once here
# (not on every expansion); git or tags missing yields an empty value silently.
ifeq ($(origin VERSION),undefined)
VERSION := $(patsubst v%,%,$(shell command -v git >/dev/null 2>&1 && git describe --tags --abbrev=0 2>/dev/null))
endif
VERSION_PKG := pixel-remastered-save-editor/browser.Version

LDFLAGS  := -s -w $(if $(VERSION),-X $(VERSION_PKG)=$(VERSION))
GOFLAGS_BUILD = -trimpath $(if $(TAGS),-tags '$(TAGS)')

LINUX_BIN   := $(DIST)/$(APP)
WINDOWS_BIN := $(DIST)/$(APP).exe

DESKTOP_FILE := packaging/linux/$(APP_ID).desktop
ICON         := icon.png
# icon.png is 256x256
ICON_SIZE    := 256x256

LINUX_PKG_NAME := $(APP)-linux-amd64
LINUX_PKG_DIR  := $(DIST)/$(LINUX_PKG_NAME)
LINUX_TARBALL  := $(DIST)/$(LINUX_PKG_NAME).tar.xz

# Asset names are deliberately version-free: GitHub release assets already live
# under their tag, and stable names make .../releases/latest/download/<name>
# links work.
WINDOWS_ZIP := $(DIST)/$(APP)-windows-amd64.zip

# Pinned for reproducible builds; bump when a new go-winres is released.
GO_WINRES_VERSION ?= v0.3.3

PREFIX  ?= $(HOME)/.local
DESTDIR ?=

# Windows resources (icon + manifest). The _windows_ suffix means Go only links
# these when GOOS=windows, so they are harmless for the Linux build.
WINRES_SYSO := rsrc_windows_amd64.syso rsrc_windows_386.syso

# go install puts binaries in $GOBIN, or $GOPATH/bin when GOBIN is unset; that
# directory is often not on PATH, so look there explicitly. Recursive (=)
# variables: only evaluated when a recipe actually needs them.
GO_INSTALL_DIR = $(or $(shell $(GO) env GOBIN 2>/dev/null),$(shell $(GO) env GOPATH 2>/dev/null)/bin)
GO_WINRES      = $(or $(shell command -v go-winres 2>/dev/null),$(GO_INSTALL_DIR)/go-winres)

ifeq ($(OS),Windows_NT)
# Native Windows build: the MinGW gcc on PATH is simply "gcc".
WINDOWS_CC ?= gcc
.DEFAULT_GOAL := build-windows
else
# Cross-compiling from Linux. Default goal is the host build only: it needs
# nothing beyond Go, gcc and the OpenGL/X11/Wayland dev libs, whereas "all" also requires
# mingw-w64, which would make a plain "make" fail on a fresh Linux box.
WINDOWS_CC ?= x86_64-w64-mingw32-gcc
.DEFAULT_GOAL := build-linux
endif

# Compress $(1) with upx when available.
define compress
@if [ -z "$(UPX)" ]; then \
		echo "upx disabled (UPX=), not compressing $(1)"; \
	elif command -v "$(UPX)" >/dev/null 2>&1; then \
		echo "$(UPX) -9 $(1)"; "$(UPX)" -9 -q "$(1)"; \
	else \
		echo "upx not found, skipping compression of $(1)"; \
	fi
endef

.PHONY: all build-linux build-windows winres setup package-linux package-windows \
	install-linux uninstall-linux clean

all: build-linux build-windows

build-linux:
	@mkdir -p $(DIST)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS_BUILD) -ldflags="$(LDFLAGS)" -o $(LINUX_BIN)
	$(call compress,$(LINUX_BIN))

build-windows: winres
	@mkdir -p $(DIST)
	CC=$(WINDOWS_CC) CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS_BUILD) -ldflags="$(LDFLAGS) -H=windowsgui" -o $(WINDOWS_BIN)
	$(call compress,$(WINDOWS_BIN))

# Only (re)generated when icon.png changes. Missing go-winres is not fatal:
# the .exe is still built, just without icon and manifest.
winres: $(firstword $(WINRES_SYSO))

$(firstword $(WINRES_SYSO)): $(ICON)
	@if [ -x "$(GO_WINRES)" ] || command -v "$(GO_WINRES)" >/dev/null 2>&1; then \
		echo "$(GO_WINRES) simply --icon $(ICON)"; \
		"$(GO_WINRES)" simply --icon $(ICON); \
	else \
		echo "warning: go-winres not found (run 'make setup'); the .exe will have no icon/manifest" >&2; \
	fi

setup:
	$(GO) install github.com/tc-hib/go-winres@$(GO_WINRES_VERSION)
	$(MAKE) --no-print-directory winres

package-linux: build-linux
	rm -rf $(LINUX_PKG_DIR)
	mkdir -p $(LINUX_PKG_DIR)
	cp $(LINUX_BIN) $(LINUX_PKG_DIR)/$(APP)
	cp $(ICON) $(LINUX_PKG_DIR)/$(APP_ID).png
	cp $(DESKTOP_FILE) $(LINUX_PKG_DIR)/$(APP_ID).desktop
	cp LICENSE README.md $(LINUX_PKG_DIR)/
	tar -C $(DIST) -cJf $(LINUX_TARBALL) $(LINUX_PKG_NAME)
	rm -rf $(LINUX_PKG_DIR)
	@echo "Created $(LINUX_TARBALL)"

# zip -j stores the files without their directory; bsdtar/python are fallbacks
# for systems without Info-ZIP (python's zipfile CLI also stores basenames).
package-windows: build-windows
	rm -f $(WINDOWS_ZIP)
	@if command -v zip >/dev/null 2>&1; then \
		echo "zip -j $(WINDOWS_ZIP) $(WINDOWS_BIN) LICENSE README.md"; \
		zip -j -q "$(WINDOWS_ZIP)" "$(WINDOWS_BIN)" LICENSE README.md; \
	elif command -v bsdtar >/dev/null 2>&1; then \
		echo "bsdtar -a -cf $(WINDOWS_ZIP) ..."; \
		bsdtar -a -cf "$(abspath $(WINDOWS_ZIP))" -C "$(abspath $(DIST))" $(APP).exe -C "$(CURDIR)" LICENSE README.md; \
	elif command -v python3 >/dev/null 2>&1; then \
		echo "python3 -m zipfile -c $(WINDOWS_ZIP) ..."; \
		python3 -m zipfile -c "$(WINDOWS_ZIP)" "$(WINDOWS_BIN)" LICENSE README.md; \
	else \
		echo "error: need zip, bsdtar or python3 to create $(WINDOWS_ZIP)" >&2; exit 1; \
	fi
	@echo "Created $(WINDOWS_ZIP)"

# Builds the binary only if it does not exist yet, so that e.g.
# "sudo make install-linux PREFIX=/usr/local" does not run go build as root.
$(LINUX_BIN):
	$(MAKE) --no-print-directory build-linux

# Exec is rewritten to an absolute path because ~/.local/bin is not always on
# the PATH of the desktop session. Menu/icon caches are refreshed only when
# DESTDIR is empty (i.e. not when staging a distro package).
install-linux: $(LINUX_BIN)
	install -Dm755 $(LINUX_BIN) "$(DESTDIR)$(PREFIX)/bin/$(APP)"
	install -Dm644 $(ICON) "$(DESTDIR)$(PREFIX)/share/icons/hicolor/$(ICON_SIZE)/apps/$(APP_ID).png"
	install -d "$(DESTDIR)$(PREFIX)/share/applications"
	sed 's|^Exec=.*|Exec="$(PREFIX)/bin/$(APP)"|' $(DESKTOP_FILE) > "$(DESTDIR)$(PREFIX)/share/applications/$(APP_ID).desktop"
	chmod 644 "$(DESTDIR)$(PREFIX)/share/applications/$(APP_ID).desktop"
	-@[ -z "$(DESTDIR)" ] && command -v update-desktop-database >/dev/null 2>&1 && update-desktop-database -q "$(DESTDIR)$(PREFIX)/share/applications" || true
	-@[ -z "$(DESTDIR)" ] && command -v gtk-update-icon-cache >/dev/null 2>&1 && gtk-update-icon-cache -q -t "$(DESTDIR)$(PREFIX)/share/icons/hicolor" 2>/dev/null || true
	@echo "Installed $(APP) to $(DESTDIR)$(PREFIX)"

uninstall-linux:
	rm -f "$(DESTDIR)$(PREFIX)/bin/$(APP)" \
		"$(DESTDIR)$(PREFIX)/share/icons/hicolor/$(ICON_SIZE)/apps/$(APP_ID).png" \
		"$(DESTDIR)$(PREFIX)/share/applications/$(APP_ID).desktop"

clean:
	rm -rf $(DIST) $(WINRES_SYSO)
