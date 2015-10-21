libsass
=========

[![Circle CI](https://circleci.com/gh/wellington/go-libsass.svg?style=svg)](https://circleci.com/gh/wellington/go-libsass) [![Build status](https://ci.appveyor.com/api/projects/status/uhl4swbb2r7lcfpc/branch/master?svg=true)](https://ci.appveyor.com/project/drewwells/go-libsass/branch/master)



Libsass provides Go bindings for the speedy [sass/libsass](http://github.com/sass/libsass) project.

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
