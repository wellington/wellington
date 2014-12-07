FROM golang:1.4rc2
WORKDIR /usr/src/myapp
ADD . /usr/src/myapp

ENV GOPATH /usr

# install g++
RUN apt-get update
RUN apt-get -y install g++ graphviz

RUN make deps
RUN go get ./...

VOLUME ["/gopath/src/myapp","/rmn"]

CMD []
