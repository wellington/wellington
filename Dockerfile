FROM gliderlabs/alpine:latest

# install g++
RUN apk update
RUN apk add build-base pkgconf autoconf automake libtool git file mercurial
RUN curl -LsO https://circle-artifacts.com/gh/andyshinn/alpine-pkg-go/2/artifacts/0/home/ubuntu/alpine-pkg-go/packages/x86_64/go-1.4.2-r0.apk && apk --allow-untrusted add --update go-1.4.2-r0.apk
ENV GOPATH /usr
ENV GOROOT /usr/lib/go
RUN go version

ENV libsass_ver a73ae2637e2a004f98959d28b39fe073125000af
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
RUN make godeps
RUN cd wt && go install
#RUN cd wt && godep go install
