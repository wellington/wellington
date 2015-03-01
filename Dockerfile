FROM gliderlabs/alpine

# install g++
RUN apk update
RUN apk add go build-base pkgconf autoconf automake libtool git

ENV libsass_ver 3.1.0
ENV LIBSASSPATH /build/libsass
ENV GOPATH /usr/src

ADD https://github.com/sass/libsass/archive/$libsass_ver.tar.gz /usr/src/libsass.tar.gz
RUN tar xvzf /usr/src/libsass.tar.gz -C /usr/src

WORKDIR /usr/src/libsass-$libsass_ver
#RUN find .
#ENV BUILD static
RUN autoreconf --force --install
RUN ./configure --disable-tests --disable-shared \
             --prefix=$LIBSASSPATH --disable-silent-rules \
			 --disable-dependency-tracking
RUN make install

COPY . /usr/src/app
COPY . /usr/src/github.com/wellington/wellington
# Inject Godep
#ADD https://github.com/tools/godep/archive/master.tar.gz /usr/src/godep.tar.gz

ENV PKG_CONFIG_PATH $LIBSASSPATH/lib/pkgconfig
#WORKDIR /usr/src/app
#RUN make deps #inlined this command to speed up docker build
#RUN go get -d -v ./...
RUN cd $GOPATH/github.com/wellington/wellington/wt && go get .
RUN find $GOPATH
