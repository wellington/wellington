FROM gliderlabs/alpine:3.2

# install g++
RUN apk --update add build-base pkgconf autoconf automake libtool git file mercurial go

ENV GOPATH /usr
ENV GOROOT /usr/lib/go
RUN go version

ENV LIBSASSPATH /build/libsass
ENV PKG_CONFIG_PATH $LIBSASSPATH/lib/pkgconfig
ENV GOPATH /usr

COPY . /usr/src/github.com/wellington/wellington
RUN go get github.com/tools/godep
WORKDIR /usr/src/github.com/wellington/wellington
RUN $GOPATH/bin/godep restore

WORKDIR /usr/src/github.com/wellington/wellington
RUN make install
