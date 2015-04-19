#!/bin/bash
FOLDER=libsass-src
HASH=$(cat .libsass_version)
export LIBSASS_VERSION=$HASH
[ -d $FOLDER ] || mkdir $FOLDER
cd $FOLDER
if [ -f "$FOLDER/libsass.$HASH.tar.gz" ];
then
	echo "Cache found $HASH"
else
	echo "Fetching source of https://github.com/sass/libsass/archive/$HASH.tar.gz"
	#make clean
	curl -s -S -L "https://github.com/sass/libsass/archive/$HASH.tar.gz" -o "libsass.$HASH.tar.gz"
	tar xvf "libsass.$HASH.tar.gz" --strip 1
fi

autoreconf -fvi
./configure --disable-shared --prefix=$(pwd) --disable-silent-rules --disable-dependency-tracking
# Check file permissions
make install
