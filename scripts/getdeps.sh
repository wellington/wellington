#!/bin/sh

HASH="85bccf28544106546aa5e1cae78bdaf7421b70dc"

[ -d libsass ] || mkdir libsass

cd libsass
curl -L "https://github.com/sass/libsass/archive/$HASH.tar.gz" -o libsass.tar.gz
tar xvf libsass.tar.gz --strip 1
sudo make install
cd ..
