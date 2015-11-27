FROM gliderlabs/alpine:latest

COPY lib64 /usr/lib

COPY wt /usr/bin/wt
# commenting this line cut the image size in half
# RUN chmod +x /usr/bin/wt

EXPOSE 12345

VOLUME "/data"
WORKDIR /data

CMD wt compile --images-dir /data/img -b /data/build --gen /data/build/img /data/sass
