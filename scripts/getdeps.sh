#!/bin/sh

HASH=$(cat libsass/.lib_version)

[ -d libsass ] || mkdir libsass
if [ -f libsass/"libsass.$HASH.tar.gz" ];
then
	echo "Cache found $HASH"
else
	echo "Fetching source of $HASH"
	cd libsass
	make clean
	curl -L "https://github.com/drewwells/libsass/archive/$HASH.tar.gz" -o "libsass.$HASH.tar.gz"
	tar xvf "libsass.$HASH.tar.gz" --strip 1
	cd ..
fi
# Check file permissions
if [ -w '/usr/local/lib/libsass.a' ];
then
	cd libsass; make clean install
else
	cd libsass; sudo make clean install
fi
