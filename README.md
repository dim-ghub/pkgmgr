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

Sources: `pacman`, `aur`, `flatpak`, `gh` (e.g. `pkg install flatpak/org.mozilla.firefox --user`)

## GitHub plugins (`gh`)

Packages hosted on GitHub can be defined in the embedded default config or
overridden in `~/.config/pkgmgr/plugins.json`. The binary ships with package
definitions for [dim-ghub](https://github.com/dim-ghub)'s forks:

```
pkg search caelestia
pkg install gh/dim-ghub/caelestia-shell-git
pkg install gh/dim-ghub/caelestia-cli-git
```

Each plugin specifies a GitHub repo, clone destination, and the shell commands
to install/update. The manager clones the repo if needed, pulls updates, and
runs the steps. Install status checks the clone directory first, then falls
back to an optional `checkBin` binary path (used for self-detection).

pkgmgr can self-update via `pkg update` (the binary includes itself in the
default plugin config):

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
