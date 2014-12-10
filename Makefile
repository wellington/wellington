.PHONY: test
current_dir = $(shell pwd)
rmnpath = "/Users/drew/work/rmn"
guipath = $(rmnpath)/www/gui
FILES := $(shell find $(rmnpath)/www/gui/sass -name "*.scss")
echo:
	echo $(current_dir)
install:
	go install github.com/wellington/wellington/wt
home:
	go run wpt/main.go -gen ~/work/rmn/www/gui/build/im -b ~/work/rmn/www/gui/build/css/ -p ~/work/rmn/www/gui/sass -d ~/work/rmn/www/gui/im/sass ~/work/rmn/www/gui/sass/_pages/home.scss
profile: install
	wt --cpuprofile=wt.prof -gen $(guipath)/build/im -font $(guipath)/font-face -b $(guipath)/build/css/ -p $(guipath)/sass -d $(guipath)/im/sass $(FILES)
	go tool pprof --png $(GOPATH)/bin/wt wt.prof > profile.png
	open profile.png
deps:
	scripts/getdeps.sh
headers:
	scripts/getheaders.sh
build:
	docker build -t wellington .
docker:
	docker run -it -v $(rmnpath):/rmn -v $(current_dir):/usr/src/myapp wellington bash
dockerprofile:
	docker run -it -v $(rmnpath):/rmn -v $(current_dir):/usr/src/myapp wellington make dockerexec
dockerexec:
	go run sprite/main.go -gen /rmn/www/gui/build/im -font /rmn/www/gui/font-face  --cpuprofile=sprite.prof -b /rmn/www/gui/build/css/ -p /rmn/www/gui/sass -d /rmn/www/gui/im/sass /rmn/www/gui/**/*.scss
	go tool pprof --pdf /usr/bin/sprite sprite.prof > profile.pdf
test:
	scripts/goclean.sh
compass:
	cd ~/work/rmn && grunt clean && time grunt build_css
swift: install
	time wt -gen $(guipath)/build/im -font $(guipath)/font-face -b $(guipath)/build/css/ -p $(guipath)/sass -d $(guipath)/im/sass $(FILES)
time: compass swift
