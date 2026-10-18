# Final Fantasy VI Save Editor

Supports editing `Final Fantasy VI Pixel Remastered` Save files.
- Character Stats, Magic, Abilities, Equipment, Commands, and Jobs
- Inventory and Important items
- Change Party members
- Map data including player world/map/position/facing, Black Jack world/position/facing, and Falcon world/position/facing.
- Bestiary and ability to mark every mob as being defeated

Known Issue:
Better multi-lingual support is coming with an update for the GUI API. For now, if a certain language is needed, download a TTF of that language and place it in the same directory as `FFPR_Save_Editor.exe` (or point the `FYNE_FONT` environment variable at it, e.g. `FYNE_FONT=/path/to/font.ttf`) and then start the application.
Google search with `"free [language] fonts"` for TTF downloads. Replace `[lanugage]` with what you need, ie: `"free japanese fonts"`

![main](https://github.com/KiameV/final-fantasy-pr-save-editor/assets/26260306/091d8b7a-f445-4526-967b-48df49cbc459)
![1](https://github.com/KiameV/final-fantasy-pr-save-editor/assets/26260306/dcfcf3cd-5662-444f-b2d8-7d69a849a148)
![2](https://github.com/KiameV/final-fantasy-pr-save-editor/assets/26260306/079eb6c4-3462-4eca-8c0f-b6248fdd8a9f)
![3](https://github.com/KiameV/final-fantasy-pr-save-editor/assets/26260306/4dd753f7-d227-49c7-9b38-359eb6c7409b)

## Running on Linux

A native Linux build is available (see [Building](#building)). Saves of the Steam/Proton version of the games live in the Proton prefix:

```
~/.local/share/Steam/steamapps/compatdata/<appid>/pfx/drive_c/users/steamuser/Documents/My Games/<GAME> PR/Steam/<SteamID64>/
```

| Game | Steam app ID |
|------|--------------|
| Final Fantasy I PR   | 1173770 |
| Final Fantasy II PR  | 1173780 |
| Final Fantasy III PR | 1173790 |
| Final Fantasy IV PR  | 1173800 |
| Final Fantasy V PR   | 1173810 |
| Final Fantasy VI PR  | 1173820 |

The editor tries to detect these folders automatically; if it can't find yours, use the `Change` button to pick the save directory.

The configuration file is stored in `~/.config/ffpr-save-editor/` on Linux and in `%AppData%\ffpr-save-editor\` on Windows. If a `prSaveEditor.config` already sits next to the executable, that one is used instead (portable mode).

For non-Latin languages, set `FYNE_FONT=/path/to/font.ttf` when starting the editor (e.g. `FYNE_FONT=~/fonts/NotoSansJP-Regular.ttf ./FFPR_Save_Editor`), as an alternative to placing the TTF next to the executable.

## Building

### Requirements
- Go 1.23 or newer
- A C compiler (the GUI uses cgo)
- On Linux, the OpenGL, X11, Wayland and GTK3 development libraries:
  - Arch: `sudo pacman -S go gcc pkgconf gtk3 libxcursor libxrandr libxinerama libxi libxxf86vm mesa wayland libxkbcommon`
    - plus `mingw-w64-gcc` to cross-compile the Windows build, and optionally `upx`
  - Debian/Ubuntu: `sudo apt install golang gcc pkg-config libgl1-mesa-dev xorg-dev libgtk-3-dev libwayland-dev libxkbcommon-dev`
    - plus `gcc-mingw-w64-x86-64` to cross-compile the Windows build, and optionally `upx`
  - Fedora: `sudo dnf install golang gcc pkgconf-pkg-config mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel gtk3-devel wayland-devel libxkbcommon-devel`
    - plus `mingw64-gcc` to cross-compile the Windows build, and optionally `upx`

### Commands
All outputs go to `dist/`.

| Command | What it does |
|---------|--------------|
| `make setup` | Installs `go-winres` and generates the Windows icon/manifest resources |
| `make build-linux` (default) | Builds `dist/FFPR_Save_Editor` |
| `make build-windows` | Cross-compiles `dist/FFPR_Save_Editor.exe` (needs mingw-w64) |
| `make all` | Both of the above |
| `make package-linux` | Creates `dist/FFPR_Save_Editor-linux-amd64.tar.xz` (binary, icon, `.desktop` file) |
| `make install-linux` | Installs the binary, icon and menu entry into `~/.local` (override with `PREFIX=...`) |
| `make uninstall-linux` | Removes what `install-linux` installed |
| `make clean` | Removes `dist/` and the generated Windows resources |

If `upx` is installed the binaries are compressed with it; otherwise that step is skipped (`make UPX=` disables it explicitly).

The Linux binary runs natively on Wayland when available and falls back to X11 otherwise. Under Wayland the window icon comes from the installed `.desktop` file (`make install-linux`).

### Building natively on Windows
Install Go and a MinGW-w64 gcc (e.g. [MSYS2](https://www.msys2.org/) or TDM-GCC), make sure `gcc` is on your `PATH`, then:

```
go install github.com/tc-hib/go-winres@latest
go-winres simply --icon icon.png
go build -ldflags="-s -w -H=windowsgui" -o FFPR_Save_Editor.exe
```

From an MSYS2 shell with `make` installed, running `make` does the same and writes `dist/FFPR_Save_Editor.exe`.
