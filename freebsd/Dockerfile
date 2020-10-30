FROM egypcio/freebsd:11.0-release

RUN mount -t fdescfs fdesc /dev/fd
RUN pkg install -y bash

ADD overrides.scss /sass/overrides.scss
ADD wt /usr/bin/wt
