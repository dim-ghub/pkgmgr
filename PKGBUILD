# Maintainer: dim-ghub <mlrudasill@gmail.com>

pkgname=pkgmgr-git
pkgver=0.1.0
pkgrel=1
pkgdesc='Unified CLI wrapper for pacman, AUR (paru), and Flatpak'
arch=('x86_64')
url='https://github.com/dim-ghub/pkgmgr'
license=('MIT')
makedepends=('go' 'git')
depends=('flatpak')
optdepends=('paru: AUR support')
source=("$pkgname::git+https://github.com/dim-ghub/pkgmgr.git")
sha256sums=('SKIP')

pkgver() {
	cd "$srcdir/$pkgname"
	git describe --long --tags 2>/dev/null || echo "0.1.0"
}

build() {
	cd "$srcdir/$pkgname"
	go build -o pkg .
}

package() {
	cd "$srcdir/$pkgname"
	install -Dm755 pkg "$pkgdir/usr/bin/pkg"
	install -Dm644 README.md "$pkgdir/usr/share/doc/pkgmgr/README.md"
	install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}
