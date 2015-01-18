FROM golang:1.4rc2

# install g++
RUN apt-get update
RUN apt-get -y install g++ pkg-config dh-autoreconf

ENV libsass_ver 3.1.0

RUN curl -sSL https://github.com/sass/libsass/archive/$libsass_ver.tar.gz \
		| tar -v -C /usr/src -xz

WORKDIR /usr/src/libsass-$libsass_ver
RUN autoreconf --force --install
RUN ./configure --disable-tests --disable-shared \
             --prefix=/usr --disable-silent-rules \
			 --disable-dependency-tracking
RUN make install

COPY . /usr/src/app
COPY . /go/src/github.com/wellington/wellington

ENV PKG_CONFIG_PATH /usr/lib/pkgconfig
WORKDIR /usr/src/app
#RUN make deps #inlined this command to speed up docker build
RUN go get -d -v ./...
RUN go install github.com/wellington/wellington/wt

EXPOSE 12345
VOLUME "/data"

WORKDIR /data
#CMD [ "sh", "-c", "echo", "$HOME" ]
CMD [ "wt", "-http", "-p", "/data", "-d", "/data/img", "-b", "/data", "-gen", "/data/build" ]
