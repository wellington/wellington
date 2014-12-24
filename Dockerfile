FROM golang:1.4rc2

# install g++
RUN apt-get update
RUN apt-get -y install g++ pkg-config dh-autoreconf

COPY . /usr/src/app

ENV PKG_CONFIG_PATH /root/lib/pkgconfig
WORKDIR /usr/src/app
RUN make deps
RUN go get -d -v ./...
RUN ln -s /usr/src/myapp /go/src/github.com/wellington/wellington
RUN make install
EXPOSE 12345
VOLUME "/data"

CMD []
