#!/bin/bash

HASH=$(cat .libsass_version)
export LIBSASS_VERSION=HASH
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
touch /usr/local/lib/libsass.a
if [ -w '/usr/local/lib/libsass.a' ];
then
	cd libsass
	autoreconf -fvi
	./configure --disable-silent-rules --disable-dependency-tracking --enable-static
	make install
	#delete shared libraries if found
	rm /usr/local/lib/libsass.so
else
	cd libsass
	autoreconf -fvi
	./configure --disable-silent-rules --disable-dependency-tracking --enable-static
	sudo make install
	#delete shared libraries if found
	sudo rm /usr/local/lib/libsass.so
fi
