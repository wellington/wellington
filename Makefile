.PHONY: test

home:
	go run sprite/main.go -gen ~/work/rmn/www/gui/build/im -b ~/work/rmn/www/gui/build/css/ -p ~/work/rmn/www/gui/sass -d ~/work/rmn/www/gui/im/sass ~/work/rmn/www/gui/sass/_pages/home.scss
profile:
	go run sprite/main.go -gen ~/work/rmn/www/gui/build/im  --cpuprofile=sprite.prof -b ~/work/rmn/www/gui/build/css/ -p ~/work/rmn/www/gui/sass -d ~/work/rmn/www/gui/im/sass ~/work/rmn/www/gui/**/*.scss
	go tool pprof $(which sprite) sprite.prof
deps:
	scripts/getdeps.sh
headers:
	scripts/getheaders.sh
build:
	docker build -t sprite .
exec:
	docker run -it -v "$(pwd)":/gopath/src/app sprite bash
test:
	scripts/goclean.sh
compass:
	cd ~/work/rmn && grunt clean && time grunt build_css
swift:
	time go run sprite/main.go -gen ~/work/rmn/www/gui/build/im -b ~/work/rmn/www/gui/build/css/ -p ~/work/rmn/www/gui/sass -d ~/work/rmn/www/gui/im/sass ~/work/rmn/www/gui/**/*.scss
time: compass swift
