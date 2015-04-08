FROM gliderlabs/alpine:edge

# install g++
RUN apk --update add build-base pkgconf autoconf automake libtool git file mercurial go

ENV GOPATH /usr
RUN go version

ENV libsass_ver d215db5edb90035d18b616a499730841a5b622df
ENV LIBSASSPATH /build/libsass
ENV PKG_CONFIG_PATH $LIBSASSPATH/lib/pkgconfig
ENV GOPATH /usr

ADD https://github.com/sass/libsass/archive/$libsass_ver.tar.gz /usr/src/libsass.tar.gz
RUN tar xzf /usr/src/libsass.tar.gz -C /usr/src

WORKDIR /usr/src/libsass-$libsass_ver

RUN autoreconf -fvi
RUN ./configure --disable-tests --disable-shared \
             --prefix=$LIBSASSPATH --disable-silent-rules \
			 --disable-dependency-tracking
RUN make install

COPY . /usr/src/app
COPY . /usr/src/github.com/wellington/wellington

WORKDIR /usr/src/app

RUN make godep
RUN cd wt && go install
#RUN cd wt && godep go install
