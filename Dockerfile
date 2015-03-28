FROM gliderlabs/alpine:latest

# install g++
RUN apk update
RUN apk add go build-base pkgconf autoconf automake libtool git

ENV libsass_ver 8e7a2947b
ENV LIBSASSPATH /build/libsass
ENV GOPATH /usr/src

ADD https://github.com/sass/libsass/archive/$libsass_ver.tar.gz /usr/src/libsass.tar.gz
RUN tar xvzf /usr/src/libsass.tar.gz --strip 1 -C /usr/src/${libsass_ver}

WORKDIR /usr/src/libsass-${libsass_ver}

RUN autoreconf -fvi
RUN ./configure --disable-tests --disable-shared \
             --prefix=$LIBSASSPATH --disable-silent-rules \
			 --disable-dependency-tracking
RUN make install

COPY . /usr/src/app
COPY . /usr/src/github.com/wellington/wellington

ENV PKG_CONFIG_PATH $LIBSASSPATH/lib/pkgconfig
RUN cd $GOPATH/github.com/wellington/wellington/wt && go get .
RUN find $GOPATH
