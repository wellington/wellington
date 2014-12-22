#!/bin/bash

HASH=$(cat .libsass_version)
export LIBSASS_VERSION=$HASH
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
cd libsass
autoreconf -fvi
./configure --disable-shared --prefix=$HOME --disable-silent-rules --disable-dependency-tracking
# Check file permissions
touch /usr/local/lib/libsass.a
if [ -w '/usr/local/lib/libsass.a' ];
then
	make install
else
	sudo make install
fi
