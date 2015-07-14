libsass
=========

Libsass provides Go bindings for the speedy [sass/libsass](github.com/sass/libsass) project.

To build, setup Go

    go build

To test

    go test

### FAQ

Rebuilding libsass on each go compilation is very slow. Optionally, it is
possible to link against a system installed libsass. To leverage a system
installed libsass, use `-tags dev`.

    go build -tags dev
    go test -tags dev
