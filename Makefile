.PHONY: test build
current_dir = $(shell pwd)
rmnpath = $(RMN_BASE_PATH)
guipath = $(rmnpath)/www/gui
libsass_ver = $(shell cat \.libsass_version)
wt_ver = $(shell cat version.txt)
LASTGOPATH=$(shell python -c "import os; a=os.environ['GOPATH']; print a.split(':')[-1]")

export PKG_CONFIG_PATH=$(current_dir)/../go-libsass/lib/pkgconfig

install: deps
	go install -ldflags "-X github.com/wellington/wellington/version.Version $(wt_ver)" github.com/wellington/wellington/wt

deps: godep
	godep restore
	# [ -d ../go-libsass ] ||	go get github.com/wellington/go-libsass
	# cd ../go-libsass && ls -l #without this like, $(MAKE) fails, go figure?
	cd ../go-libsass && $(MAKE) deps

$(LASTGOPATH)/bin/goxc:
	go get github.com/laher/goxc

release: $(LASTGOPATH)/bin/goxc
	goxc -tasks='xc archive' -build-ldflags "-X github.com/wellington/wellington/version.Version $(wt_ver)" -bc='darwin' -arch='amd64' -wd=wt -d=snapshot -pv $(wt_ver) -n wt

windows:
	go build -o wt.exe -x -ldflags "-extldflags '-static' -X=github.com/wellington/wellington/version.Version=$(wt_ver)" github.com/wellington/wellington/wt

bench:
	go test ./... -bench=.
home:
	go run wt/main.go -font $(guipath)/font-face -gen $(guipath)/build/im -b $(guipath)/build/css/ -p $(guipath)/sass -d $(guipath)/im/sass $(guipath)/sass/_pages/home.scss
homewatch:
	go run wt/main.go --watch -font $(guipath)/font-face -gen $(guipath)/build/im -b $(guipath)/build/css/ -p $(guipath)/sass -d $(guipath)/im/sass $(guipath)/sass/_pages/home.scss

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
	#chown -R $(EUID):$(EGID) /build/libsass
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

NONVENDOR = $(shell go list ./... | grep -v /vendor/ | grep -v /examples/)
DIRS = $(shell go list -f '{{.Dir}}' ./... | grep -v /vendor/)

.PHONY: gover.coverprofile
gover.coverprofile:
	go get golang.org/x/tools/cmd/vet
	# retrieve lint and test deps
	go get github.com/axw/gocov/gocov
	go get github.com/mattn/goveralls
	go get github.com/modocache/gover
	go get golang.org/x/tools/cmd/goimports
	go get github.com/golang/lint/golint
	go get golang.org/x/tools/cmd/cover
	go list -f 'golint {{.Dir}}' | xargs -L 1 sh -c
	go list -f '{{if len .TestGoFiles}}"go test -covermode=count -short -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"{{end}}' ./... | xargs -L 1 sh -c
	gover

godeptest:
	godep go test -i -v $(NONVENDOR)
	godep go test -race -i -v $(NONVENDOR)
	go list -f '{{if len .TestGoFiles}}"godep go test -v -race {{.ImportPath}}"{{end}}' ./... | xargs -L 1 sh -c

test: godep godeptest

cover: gover.coverprofile

compass:
	cd ~/work/rmn && grunt clean && time grunt build_css
save:
	cd libsass-src; git rev-parse HEAD > ../.libsass_version
profile: #install
	scripts/profile.sh
swift: install
	scripts/swift.sh
watch: install
	scripts/watch.sh
time: compass swift
