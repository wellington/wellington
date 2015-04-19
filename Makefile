.PHONY: test build
current_dir = $(shell pwd)
rmnpath = $(RMN_BASE_PATH)
guipath = $(rmnpath)/www/gui
libsass_ver = $(shell cat \.libsass_version)
LASTGOPATH=$(shell python -c "import os; a=os.environ['GOPATH']; print a.split(':')[-1]")

ifndef PKG_CONFIG_PATH
	export PKG_CONFIG_PATH=$(current_dir)/libsass-src/lib/pkgconfig
endif

pkgconfig:
	pkg-config --cflags --libs libsass

install: deps
	echo "PKG_CONFIG_PATH $(PKG_CONFIG_PATH)"
	godep go install github.com/wellington/wellington/wt

deps: libsass-src/lib/libsass.a

bench:
	go test ./... -bench=.
home:
	go run wt/main.go -font $(guipath)/font-face -gen $(guipath)/build/im -b $(guipath)/build/css/ -p $(guipath)/sass -d $(guipath)/im/sass $(guipath)/sass/_pages/home.scss
homewatch:
	go run wt/main.go --watch -font $(guipath)/font-face -gen $(guipath)/build/im -b $(guipath)/build/css/ -p $(guipath)/sass -d $(guipath)/im/sass $(guipath)/sass/_pages/home.scss

profile: install
	wt --cpuprofile=wt.prof -gen $(guipath)/build/im -font $(guipath)/font-face -b $(guipath)/build/css/ -p $(guipath)/sass -d $(guipath)/im/sass $(FILES)
	go tool pprof --png $(GOPATH)/bin/wt wt.prof > profile.png
	open profile.png

$(LASTGOPATH)/bin/godep:
	go get github.com/tools/godep

godep: $(LASTGOPATH)/bin/godep

libsass-src/*:
	mkdir -p libsass-src

libsass-src/lib/libsass.a: libsass-src/*
	scripts/getdeps.sh

headers:
	scripts/getheaders.sh

clean:
	rm -rf build/*

copyout:
	chown $(EUID):$(EGID) $(GOPATH)/bin/wt
	cp $(GOPATH)/bin/wt /tmp
	chown -R $(EUID):$(EGID) /build/libsass
	mkdir -p /tmp/lib64
	cp /usr/lib/libstdc++.so.6 /tmp/lib64
	cp /usr/lib/libgcc_s.so.1 /tmp/lib64
	chown -R $(EUID):$(EGID) /tmp

container-build:
	- mkdir build
	- rm profile.cov
	docker build -t wt-build .
	docker run -v $(PWD)/build:/tmp -e EUID=$(shell id -u) -e EGID=$(shell id -g) wt-build make copyout

build: container-build
	cp Dockerfile.scratch build/Dockerfile
	cd build; docker build -t drewwells/wellington .

push: build
	docker push drewwells/wellington:latest
docker:
	docker run -e HOST=http://$(shell boot2docker ip):8080 -it -p 8080:12345 -v $(current_dir):/usr/src/myapp -v $(current_dir)/test:/data drewwells/wellington

profile.cov:
	go get golang.org/x/tools/cmd/vet
	# retrieve lint and test deps
	go get github.com/axw/gocov/gocov
	go get github.com/mattn/goveralls
	go get golang.org/x/tools/cmd/goimports
	go get github.com/golang/lint/golint
	go get golang.org/x/tools/cmd/cover
	scripts/goclean.sh

test: godep profile.cov

compass:
	cd ~/work/rmn && grunt clean && time grunt build_css
save:
	cd libsass-src; git rev-parse HEAD > ../.libsass_version
swift: install
	scripts/swift.sh
watch: install
	scripts/watch.sh
time: compass swift
