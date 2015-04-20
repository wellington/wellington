export PKG_CONFIG_PATH=$(shell pwd)/lib/pkgconfig

deps:
	cd libsass-src; autoreconf -fvi && \
		./configure --disable-shared --prefix=$(shell pwd) --disable-silent-rules --disable-dependency-tracking && \
		make install
.PHONY: test
test:
	@echo $(PKG_CONFIG_PATH)
	go test -race .
