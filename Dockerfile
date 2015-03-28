FROM gliderlabs/alpine:latest

# install g++
RUN apk update
RUN apk add go build-base pkgconf autoconf automake libtool git

ENV libsass_ver 8e7a2947b82adcb79484cbc0843979038c9d7c4a
ENV LIBSASSPATH /build/libsass
ENV PKG_CONFIG_PATH $LIBSASSPATH/lib/pkgconfig
ENV GOPATH /usr

ADD https://github.com/sass/libsass/archive/$libsass_ver.tar.gz /usr/src/libsass.tar.gz
RUN tar xvzf /usr/src/libsass.tar.gz -C /usr/src

WORKDIR /usr/src/libsass-${libsass_ver}

RUN autoreconf -fvi
RUN ./configure --disable-tests --disable-shared \
             --prefix=$LIBSASSPATH --disable-silent-rules \
			 --disable-dependency-tracking
RUN make install

COPY . /usr/src/app
COPY . /usr/src/github.com/wellington/wellington

WORKDIR /usr/src/app

RUN go get -d -v ./...
RUN cd wt && go install
#RUN cd wt && godep go install
