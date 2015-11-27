FROM gliderlabs/alpine:edge

# install g++
RUN apk --update add git mercurial go
# RUN apk --update add build-base pkgconf autoconf automake libtool git file mercurial go

ENV GO15VENDOREXPERIMENT 1
ENV GOPATH /usr
ENV GOROOT /usr/lib/go
RUN go version

ENV LIBSASSPATH /build/libsass
ENV PKG_CONFIG_PATH $LIBSASSPATH/lib/pkgconfig
ENV GOPATH /usr

COPY . /usr/src/github.com/wellington/wellington
RUN go get github.com/tools/godep
WORKDIR /usr/src/github.com/wellington/wellington
#RUN $GOPATH/bin/godep restore

RUN make install
