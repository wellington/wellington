#!/bin/bash

HASH=$(cat .libsass_version)
export LIBSASS_VERSION=$HASH
[ -d libsass ] || mkdir libsass
if [ -f libsass/"libsass.$HASH.tar.gz" ];
then
	echo "Cache found $HASH"
else
	echo "Fetching source of https://github.com/sass/libsass/archive/$HASH.tar.gz"
	cd libsass
	#make clean
	curl -s -S -L "https://github.com/sass/libsass/archive/$HASH.tar.gz" -o "libsass.$HASH.tar.gz"
	tar xvf "libsass.$HASH.tar.gz" --strip 1
	cd ..
fi
cd libsass
autoreconf -fvi
./configure --disable-shared --prefix=$(pwd) --disable-silent-rules --disable-dependency-tracking
# Check file permissions
make install

# temporary fix waiting for PR#998
cp sass_version.h include/sass_version.h
