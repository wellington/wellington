libsass
=========

Libsass provides Go bindings for the terrific [sass/libsass](github.com/sass/libsass) project.

To build this package, refer to the circle.yml config.

Libsass is linked to Go via pkg-config. For development, it is recommended to set your PKG_CONFIG_PATH to this:

    export PKG_CONFIG_PATH=$(pwd)/lib/pkgconfig


### Building From Source (OS X edition)

Install the following dependencies for compiling libsass:

    brew install go autoconf automake libtool mercurial pkg-config
    make deps test # if this all works, proceed to next step
    
    export PKG_CONFIG_PATH=$(GOPATH)/src/github.com/wellington/libsass/lib/pkgconfig
    go test #go commands will work once PKG_CONFIG_PATH is set


Or, you can install libsass with brew and set pkgconfig to find it. It is recommended to use head until libsass becomes more stable.

    brew install --head libsass
    export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
