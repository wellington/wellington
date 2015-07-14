export PKG_CONFIG_PATH=$(shell pwd)/lib/pkgconfig

SOURCES=libsass-src/*.cpp libsass-src/*.c libsass-src/*.h libsass-src/*.hpp
CPSOURCES=libsass-build/*.cpp libsass-build/*.c libsass-build/*.h libsass-build/*.hpp

install: deps

deps:

print-%  : ; @echo $* = $($*)

fetch:
	git submodule sync
	git submodule update --init

version:
	# hack to temporarily fix versioning
	cp libs/sass_version.h libsass-src/sass_version.h

libsass-src: fetch

libsass-tmp: clean libsass-src $(SOURCES)
	# generate configure scripts
	- cd libsass-src; autoreconf -fvi
	mkdir -p libsass-tmp
	# configure and install libsass
	cd libsass-tmp; \
		../libsass-src/configure --disable-shared \
			--prefix=$(shell pwd) --disable-silent-rules \
			--disable-dependency-tracking

.PHONY: libsass-build
libsass-build:

	# copy the updated source files into libsass-build
	mkdir -p libsass-build
	rm $(CPSOURCES)
	cp $(SOURCES) libsass-build
	#VERSION = $(shell cd libsass-src; ./version.sh)
	#manually update the version

lib: libsass-tmp
	mv libsass-src/sass_version.h libsass-src/sass_version.hold
	cp libsass-tmp/sass_version.h libsass-src/sass_version.h
	- cd libsass-tmp && make install
	mv libsass-src/sass_version.hold libsass-src/sass_version.h

.PHONY: test
test:
	go test -tags dev -race .

cleanfiles:
	rm -rf lib include libsass-src libsass-tmp

clean: cleanfiles
	git submodule update
