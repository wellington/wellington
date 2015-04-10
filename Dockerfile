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

ENV libsass_ver 3.2.0-beta.5
ENV LIBSASSPATH /build/libsass
ENV PKG_CONFIG_PATH $LIBSASSPATH/lib/pkgconfig
ENV GOPATH /usr

ADD https://github.com/sass/libsass/archive/$libsass_ver.tar.gz /usr/src/libsass.tar.gz
RUN tar xzf /usr/src/libsass.tar.gz -C /usr/src
RUN mv /usr/src/libsass-$libsass_ver /usr/src/libsass

WORKDIR /usr/src/libsass

RUN autoreconf -fvi
RUN ./configure --disable-tests --disable-shared \
                --prefix=$LIBSASSPATH --disable-silent-rules \
                --disable-dependency-tracking
RUN make && make install
COPY . /usr/src/github.com/wellington/wellington

WORKDIR /usr/src/github.com/wellington/wellington
RUN cp -R Godeps/_workspace/src/* $GOPATH/src
RUN go install github.com/wellington/wellington/wt
