FROM golang:1.4rc2

ENV libsass_ver 3.1.0-beta.2

# install g++
RUN apt-get update
RUN apt-get -y install g++ pkg-config dh-autoreconf

RUN curl -sSL https://github.com/sass/libsass/archive/$libsass_ver.tar.gz \
		| tar -v -C /usr/src -xz

WORKDIR /usr/src/libsass-$libsass_ver
RUN autoreconf --force --install
RUN ./configure --disable-tests --disable-shared \
             --prefix=/usr --disable-silent-rules \
			 --disable-dependency-tracking
RUN make install

COPY . /usr/src/app

ENV PKG_CONFIG_PATH /usr/lib/pkgconfig
WORKDIR /usr/src/app
#RUN make deps #inlined this command to speed up docker build
RUN go get -d -v ./...
RUN ln -s /usr/src/myapp /go/src/github.com/wellington/wellington
RUN make install
EXPOSE 12345
VOLUME "/data"

CMD ["wt", "-p", "/data", "-d", "/data/img", "-b", "/data/build", "-http"]
