libsass
=========

[![Circle CI](https://circleci.com/gh/wellington/go-libsass.svg?style=svg)](https://circleci.com/gh/wellington/go-libsass) [![Build status](https://ci.appveyor.com/api/projects/status/uhl4swbb2r7lcfpc/branch/master?svg=true)](https://ci.appveyor.com/project/drewwells/go-libsass/branch/master)

The only Sass compliant Go library! go-libsass is a wrapper to the [sass/libsass](http://github.com/sass/libsass) project.

To build, setup Go

    go build

To test

    go test
    
Basic example more examples found in [examples](examples)

```
buf := bytes.NewBufferString("div { p { color: red; } }")
if err != nil {
	log.Fatal(err)
}
comp, err := libsass.New(os.Stdout, buf)
if err != nil {
	log.Fatal(err)
}

if err := comp.Run(); err != nil {
	log.Fatal(err)
}
```

Output
```
div p {
  color: red; }
```

### FAQ

* Compiling go-libsass is very slow, what can be done?

    Go-libsass compiles C/C++ libsass on every build. You can install the package and speed up building `go install github.com/wellington/go-libsass`. Alternatively, it's possible to link against system libsass and forego C compiling with `go build -tags dev`.

* How do I cross compile?

    Since this package uses C bindings, you will need gcc for the target platform. For windows see, https://github.com/wellington/go-libsass/issues/37
    
    
