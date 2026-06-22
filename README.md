# pkg

Unified CLI wrapper for pacman, AUR (paru), and Flatpak.

Search, install, and remove packages from all three sources with a single command.

## Usage

```
pkg                              update all packages
pkg <query>                      search and select interactively
pkg search <query>               search and select interactively
pkg install [source/]<pkg>       install (auto-detects source)
pkg install aur/foo --noconfirm  install with flags passed through
pkg remove pacman/foo            remove (source required)
```

Sources: `pacman`, `aur`, `flatpak` (e.g. `pkg install flatpak/org.mozilla.firefox --user`)

## Install on Arch Linux

```
git clone https://github.com/dim-ghub/pkgmgr.git
cd pkgmgr
makepkg -si
```

This installs the `pkg` binary to `/usr/bin/pkg`.

## Build from source

```
go build -o pkg .
```
