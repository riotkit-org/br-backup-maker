pkgname=backup-maker
pkgver=${GITHUB_REF##*/}
pkgrel=1
pkgdesc='Tiny backup client packed in a single binary. Interacts with a `Backup Repository` server to store files, uses GPG to secure your backups even against the server administrator.'
arch=('x86_64')
url="https://github.com/riotkit-org"
license=('APACHE')
makedepends=('go')

prepare(){
    return 0
}
build() {
    echo " >> Using already built artifacts by CI from .build directory"
    return 0
}
check() {
    return 0
}
package() {
    install -Dm755 ../.build/backup-maker "$pkgdir"/usr/bin/backup-maker
    install -Dm755 ../.build/bmg "$pkgdir"/usr/bin/bmg
}
