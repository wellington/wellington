FROM subosito/golang-xc
ADD . /usr/local/go/src/github.com/drewwells/sprite_sass

WORKDIR /usr/local/go/src/github.com/drewwells/sprite_sass
RUN cd libsass; make clean all
RUN go get ./...
RUN go build

#-> % docker run -it -v /Users/drew/go/src/github.com/drewwells/sprite_sass:/usr/local/go/src/github.com/drewwells/sprite_sass subosito/golang-xc /bin/bash
