FROM golang:1.10-alpine

RUN apk add --update --no-cache alpine-sdk bash python

# compile and install librdkafka
WORKDIR /root
RUN git clone https://github.com/edenhill/librdkafka.git
WORKDIR /root/librdkafka
# checkout v0.11.6
RUN git checkout 849c066
RUN /root/librdkafka/configure
RUN make
RUN make install

# copy source files and private repo dep
COPY ./k2ws/ /go/src/k2ws/
COPY ./vendor/ /go/src/k2ws/vendor/

# static build the app
WORKDIR /go/src/k2ws
RUN go get -d ./...
RUN go build -tags=static

# create final image
FROM alpine

COPY --from=0 /go/src/k2ws/k2ws /usr/bin/

# RUN apk --no-cache add \
#       cyrus-sasl \
#       openssl \

ENTRYPOINT ["k2ws"]
