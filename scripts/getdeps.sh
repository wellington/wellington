#!/bin/sh

HASH="85bccf"

[ -d libsass ] || mkdir libsass

if [[ ! -f libsass/"libsass.$HASH.tar.gz" ]];
then
	cd libsass
	curl -L "https://github.com/sass/libsass/archive/$HASH.tar.gz" -o "libsass.$HASH.tar.gz"
	tar xvf "libsass.$HASH.tar.gz" --strip 1
	sudo make install
	cd ..
fi
