export PKG_CONFIG_PATH=$(shell pwd)/lib/pkgconfig

install: deps

deps: fetch

fetch:
	git submodule sync
	git submodule update --init

libsass-src: fetch


libsass-src/Makefile.conf: fetch

include libsass-src/Makefile.conf

LIBSASS_VERSION:=$(shell cd libsass-src; ./version.sh)
libsass-build/include/sass/version.h: libsass-src/include/sass/version.h.in
	echo "Stamping version $(LIBSASS_VERSION)"
	sed 's/@PACKAGE_VERSION@/$(LIBSASS_VERSION)/' libsass-src/include/sass/version.h.in > libsass-build/include/sass/version.h

.PHONY: libsass-build
SOURCES=libsass-build/*.cpp libsass-build/*.c libsass-build/*.h libsass-build/*.hpp
libsass-build: libsass-src
	mkdir -p libsass-build/include
	rm -rf $(SOURCES)

	# more stuff
	cp -R libsass-src/src/*.c libsass-build
	cp -R libsass-src/src/*.cpp libsass-build
	cp -R libsass-src/src/*.h libsass-build
	cp -R libsass-src/src/*.hpp libsass-build

	cp -R libsass-src/src/b64 libsass-build
	cp -R libsass-src/include libsass-build
	cp -R libsass-src/src/memory libsass-build
	cp -R libsass-src/src/utf8 libsass-build
	# hack remove the [NA] version.h
	rm libsass-build/include/sass/version.h

	touch libs/*.go

update-libsass: libsass-build libsass-build/include/sass/version.h
	@echo "Success"

.PHONY: test
test:
	go test -tags dev -race .

cleanfiles:
	rm -rf lib include libsass-src libsass-tmp

clean: cleanfiles
	git submodule update
