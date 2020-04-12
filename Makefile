.PHONY: test build
current_dir := $(shell pwd)
rmnpath := $(RMN_BASE_PATH)
guipath := $(rmnpath)/www/gui
libsass_ver := $(shell cat \.libsass_version)
wt_ver := $(shell cat version.txt)
LASTGOPATH :=$(shell echo $(GOPATH) | tr ":" "\n" | tail -1 )
HOSTOSARCH := $(shell go env GOHOSTOS)_$(shell go env GOHOSTARCH)

export PKG_CONFIG_PATH=$(current_dir)/../go-libsass/lib/pkgconfig

install:
	go install -ldflags "-X github.com/wellington/wellington/version.Version=$(wt_ver)" github.com/wellington/wellington/wt

brew: godep
	godep restore

# This is becoming a great tool for managing brew
deps: brew

$(LASTGOPATH)/bin/goxc:
	go get github.com/laher/goxc

goxc: $(LASTGOPATH)/bin/goxc

darwin:
	cd wt; go build -ldflags '-X github.com/wellington/wellington/version.Version=$(wt_ver)' -o ../snapshot/$(wt_ver)/$(HOSTOSARCH)/wt
	tar -cvzf snapshot/$(wt_ver)/wt_$(wt_ver)_$(HOSTOSARCH).tar.gz -C snapshot/$(wt_ver)/$(HOSTOSARCH)/ wt

release:
	# disabled goxc, it is incompatible with go vendoring
	#goxc -tasks='xc archive' -build-ldflags "-X github.com/wellington/wellington/version.Version $(wt_ver)" -bc='darwin' -arch='amd64' -wd=wt -d=snapshot -pv $(wt_ver) -n wt
	cd wt; go build -ldflags '-extldflags "-static" -X github.com/wellington/wellington/version.Version=$(wt_ver)' -o ../snapshot/$(wt_ver)/$(HOSTOSARCH)/wt
	tar -cvzf snapshot/$(wt_ver)/wt_$(wt_ver)_$(HOSTOSARCH).tar.gz -C snapshot/$(wt_ver)/$(HOSTOSARCH)/ wt

windows:
	go get golang.org/x/net/context
	go build -o wt.exe -x -ldflags "-extldflags '-static' -X github.com/wellington/wellington/version.Version=$(wt_ver)" github.com/wellington/wellington/wt

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
	rm -f profile.cov
	docker build --no-cache -t wt-build .
	docker run -v $(PWD)/build:/tmp -e EUID=$(shell id -u) -e EGID=$(shell id -g) wt-build sh scripts/copyout.sh
	pwd
	ls $(pwd)/build

container: container-build
	cp Dockerfile.scratch build/Dockerfile
	cd build; docker build -t drewwells/wellington .

push: container
	docker push drewwells/wellington:latest

docker:
	docker run -e HOST=http://$(shell boot2docker ip):8080 -it -p 8080:12345 -v $(current_dir):/usr/src/myapp -v $(current_dir)/test:/data drewwells/wellington

.PHONY: gover.coverprofile
gover.coverprofile:
	# retrieve lint and test deps
	go get github.com/axw/gocov/gocov
	go get github.com/mattn/goveralls
	go get github.com/modocache/gover
	go get golang.org/x/tools/cmd/cover
	go list -f '{{if len .TestGoFiles}}"go test -covermode=count -short -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"{{end}}' ./... | grep -v /vendor/ | grep -v /wt$ | xargs -L 1 sh -c
	gover . gover.coverprofile

godeptest:
	godep go test -race ./...

test:
	go test -race -v ./...

lint:
	go get github.com/golang/lint/golint
	go list -f 'golint {{.Dir}}' ./... | grep -v /vendor/ | xargs -L 1 sh -c
	go vet ./...

cover: gover.coverprofile
	go tool cover -html=gover.coverprofile

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
