libsass
=========

Libsass provides Go bindings for the speedy [sass/libsass](github.com/sass/libsass) project.

Libsass is required to use this library.

### Building From Source (OS X edition)

Install the following dependencies for compiling libsass:

    brew install go autoconf automake libtool mercurial pkg-config
    make deps test # if this all works, proceed to next step

    export PKG_CONFIG_PATH=$(GOPATH)/src/github.com/wellington/libsass/lib/pkgconfig
    go test


Or, you can install libsass with brew.

    brew install libsass
    go test
