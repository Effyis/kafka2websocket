## Kafka to WebSocket

This will create web-socket server that will serve data from a Kafka topic(s).

### config.yaml
```yaml
schema_version: "1.0"
k2ws:
  - brokers: localhost:9092
    topics: 
      - test12
      - test11m
    group_id: k2ws-test-group
    auto_offset: earliest
    auto_commit: true
    addr: :8888
    secret: onlyyouknow
  - brokers: localhost:9092
    topics: 
      - test11
    group_id: k2ws-test-group2
    auto_offset: latest
    auto_commit: false
    addr: :8889
```
This will start two web-socket servers, one on `ws://localhost:8888/onlyyouknow` and the other one on `ws://localhost:8889/`.
To test them in browser you'll have to visit `ws://localhost:8888/onlyyouknow/test` and `ws://localhost:8889/test`.

Config file must be in the same directory where executable is located and named `config.yaml`

**This project has `librdkafka` dependency.**

You can build it statically with `go build -tags static`. Check out [confluentinc/confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go#static-builds) for more info.

**Build with Docker**
```sh
docker build -t k2ws-build .
docker run --rm -v $PWD:/root/go/src/k2ws k2ws-build
```

You'll end up with `k2ws` executable that works on Ubuntu and Centos

### Test in browser

Let's assume application is running on the server `k2ws-test` and serving random topic on port `8888`.
You just need to visit `http://k2ws-test:8888/test`.
