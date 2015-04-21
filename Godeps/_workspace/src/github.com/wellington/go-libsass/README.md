libsass
=========

Libsass provides Go bindings for the terrific [sass/libsass](github.com/sass/libsass) project.

To build this package, refer to the circle.yml config.

Libsass is linked to Go via pkg-config. For development, it is recommended to set your PKG_CONFIG_PATH to this:

    export PKG_CONFIG_PATH=$(pwd)/lib/pkgconfig
