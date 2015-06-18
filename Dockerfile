FROM gliderlabs/alpine:edge

# install g++
RUN apk --update add build-base pkgconf autoconf automake libtool git file mercurial go

ENV GOPATH /usr
RUN go version

ENV LIBSASSPATH /build/libsass
ENV PKG_CONFIG_PATH $LIBSASSPATH/lib/pkgconfig
ENV GOPATH /usr

COPY . /usr/src/github.com/wellington/wellington
RUN go get github.com/tools/godep
WORKDIR /usr/src/github.com/wellington/wellington
RUN $GOPATH/bin/godep restore

WORKDIR /usr/src/github.com/wellington/go-libsass
RUN git submodule sync
RUN git submodule update --init
RUN make deps
RUN mkdir -p $LIBSASSPATH
RUN cp -R include $LIBSASSPATH
RUN cp -R lib $LIBSASSPATH

WORKDIR /usr/src/github.com/wellington/wellington
RUN make install
