export PKG_CONFIG_PATH=$(shell pwd)/lib/pkgconfig

install: deps

deps: fetch

fetch:
	git submodule sync
	git submodule update --init

libsass-src: fetch

libsass-tmp: clean libsass-src $(SOURCES)
	# generate version header from configure script
	- cd libsass-src; $(MAKE) clean; autoreconf -fvi
	mkdir -p libsass-tmp
	# configure and install libsass
	cd libsass-tmp; ../libsass-src/configure --disable-shared \
			--prefix=$(shell pwd) --disable-silent-rules \
			--disable-dependency-tracking

CPSOURCES=libsass-build/*.cpp libsass-build/*.c libsass-build/*.h libsass-build/*.hpp

libsass-src/Makefile.conf: fetch

include libsass-src/Makefile.conf

.PHONY: libsass-build
libsass-build:
	mkdir -p libsass-build/include
	rm -rf $(CPSOURCES)
	echo $(CSOURCES)
	cp -R $(addprefix libsass-src/src/,$(CSOURCES)) libsass-build
	cp -R $(addprefix libsass-src/src/,$(SOURCES)) libsass-build
	cp -R libsass-src/include libsass-build
	# more stuff
	cp -R libsass-src/src/*.hpp libsass-build
	cp -R libsass-src/src/*.h libsass-build
	cp -R libsass-src/src/b64 libsass-build
	cp -R libsass-src/src/utf8 libsass-build
	cp libsass-tmp/include/sass/version.h libsass-build/include/sass/version.h
	touch libs/*.go

copy: libsass-tmp libsass-build
	- echo "Success"

.PHONY: test
test:
	go test -tags dev -race .

cleanfiles:
	rm -rf lib include libsass-src libsass-tmp

clean: cleanfiles
	git submodule update
