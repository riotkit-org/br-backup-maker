pkgname=backup-maker
pkgver=${GITHUB_REF##*/}
pkgrel=1
pkgdesc='Tiny backup client packed in a single binary. Interacts with a `Backup Repository` server to store files, uses GPG to secure your backups even against the server administrator.'
arch=('x86_64')
url="https://github.com/riotkit-org"
license=('APACHE-2')
makedepends=('go')

prepare(){
    mkdir -p .build/
}
build() {
    cd ..
    export CGO_CPPFLAGS="${CPPFLAGS}"
    export CGO_CFLAGS="${CFLAGS}"
    export CGO_CXXFLAGS="${CXXFLAGS}"
    export CGO_LDFLAGS="${LDFLAGS}"
    export GOFLAGS="-buildmode=pie -trimpath -ldflags=-linkmode=external -mod=readonly -modcacherw"
    echo $GOROOT
    find /
    go build -o ./.build/backup-maker ./
}
check() {
    return 0
}
package() {
    install -Dm755 ../.build/backup-maker "$pkgdir"/usr/bin/$pkgname
}
