FROM centos:7

# install tools
RUN yum update && \
    yum install -y git gcc gcc-c++ make

# install golang and compile & install librdkafka
RUN cd /tmp && \
    curl -o go1.10.1.linux-amd64.tar.gz https://dl.google.com/go/go1.10.1.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.10.1.linux-amd64.tar.gz && \
    git clone https://github.com/edenhill/librdkafka.git && \
    cd /tmp/librdkafka && \
    ./configure && \
    make && \
    make install

COPY app.go config.go html-template.go k2ws.go /root/go/src/k2ws/

ENV PATH=$PATH:/usr/local/go/bin
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig

CMD cd /root/go/src/k2ws/ && \
    go get ./... && \
    go build -tags static
