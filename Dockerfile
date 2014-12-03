FROM google/golang
WORKDIR /gopath/src/app
ADD . /gopath/src/app

RUN make deps
RUN go get ./...

VOLUME ["/gopath/src/app"]

CMD []
