FROM gliderlabs/alpine:latest

# install g++
RUN apk --update add build-base pkgconf autoconf automake libtool git file mercurial
RUN curl -LsO https://circle-artifacts.com/gh/andyshinn/alpine-pkg-go/3/artifacts/0/home/ubuntu/alpine-pkg-go/packages/x86_64/go-1.4.2-r0.apk && apk --allow-untrusted add --update go-1.4.2-r0.apk

#retrieve Go source
ADD https://storage.googleapis.com/golang/go1.4.2.src.tar.gz /usr/lib/go1.4.2.src.tar.gz
RUN tar xzf /usr/lib/go1.4.2.src.tar.gz -C /usr/lib

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

WORKDIR /usr/src/github.com/wellington/go-libsass
RUN git submodule sync
RUN git submodule update --init
RUN make deps
RUN mkdir -p $LIBSASSPATH
RUN cp -R include $LIBSASSPATH
RUN cp -R lib $LIBSASSPATH

WORKDIR /usr/src/github.com/wellington/wellington
RUN make install
