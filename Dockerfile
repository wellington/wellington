FROM golang:1.4rc2
WORKDIR /usr/src/myapp

ENV GOPATH /usr

# install g++
RUN apt-get update
RUN apt-get -y install g++ graphviz pkg-config

RUN make deps
RUN go get ./...
RUN ln -s /usr/src/myapp /usr/src/github.com/wellington/wellington
VOLUME ["/usr/src/myapp","/rmn"]

CMD []
