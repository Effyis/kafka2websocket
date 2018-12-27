FROM golang:1.10-stretch

# compile & install librdkafka
RUN cd /tmp && \
    git clone https://github.com/edenhill/librdkafka.git && \
    cd /tmp/librdkafka && \
    git checkout 849c066 && \
    ./configure && \
    make && \
    make install

COPY ./k2ws/ /go/src/k2ws/
COPY ./vendor/ /go/src/k2ws/vendor/

RUN cd /go/src/k2ws/ && \
    go get ./... && \
    go build -tags static && \
    cp k2ws /usr/bin/k2ws

CMD cp /usr/bin/k2ws /build/
