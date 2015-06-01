export PKG_CONFIG_PATH=$(shell pwd)/lib/pkgconfig

install: deps

deps: fetch lib

fetch:
	git submodule sync
	git submodule update --init

libsass-build: libsass-src/*.cpp
	# generate configure scripts
	cd libsass-src; make clean && autoreconf -fvi
	- rm -rf libsass-build lib include
	mkdir -p libsass-build
	# configure and install libsass
	cd libsass-build && \
		../libsass-src/configure --disable-shared --prefix=$(shell pwd) --disable-silent-rules --disable-dependency-tracking

lib: libsass-build
	mv libsass-src/sass_version.h libsass-src/sass_version.hold
	cp libsass-build/sass_version.h libsass-src/sass_version.h
	cd libsass-build && make install
	mv libsass-src/sass_version.hold libsass-src/sass_version.h

.PHONY: test
test:
	go test -race .

clean:
	rm -rf libsass-build lib include
