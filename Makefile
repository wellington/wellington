.PHONY: test build
current_dir = $(shell pwd)
rmnpath = $(RMN_BASE_PATH)
guipath = $(rmnpath)/www/gui

echo:
	echo $(current_dir)
install:
	go get -f -u -d github.com/wellington/spritewell
	go get -f -u -d gopkg.in/fsnotify.v1
	go install github.com/wellington/wellington/wt
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
deps:
	scripts/getdeps.sh
headers:
	scripts/getheaders.sh
build:
	docker build -t drewwells/wellington .
push: build
	docker push drewwells/wellington
docker:
	docker run -it -v $(rmnpath):/rmn -v $(current_dir):/usr/src/myapp drewwells/wellington bash
test:
	scripts/goclean.sh
compass:
	cd ~/work/rmn && grunt clean && time grunt build_css
swift: install
	scripts/swift.sh
watch: install
	scripts/watch.sh
time: compass swift
