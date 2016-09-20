FROM drewwells/alpine-build:go1.7.1

RUN go version

COPY . /usr/src/github.com/wellington/wellington
WORKDIR /usr/src/github.com/wellington/wellington

RUN go get golang.org/x/net/context
RUN go install -ldflags "-X github.com/wellington/wellington/version.Version=$(cat version.txt)" github.com/wellington/wellington/wt
