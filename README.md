## Kafka to WebSocket

This will create web-socket server that will serve data from a Kafka topic(s).

### Configuration

Default configuration file is `config.yaml` located in the same directory as executable.

Example:
```yaml
schema_version: "1.0"
# cert_file: ./certs/mydomain.crt
# key_file: ./certs/mydomain.key
k2ws:
  # first Kafka config entry
  - brokers: localhost:9092
    topics: 
      - test12
      - test11m
    group_id: k2ws-test-group
    auto_offset: earliest # default is "largest"
    auto_commit: true # default is `false`
    addr: :8888
    secret: onlyyouknow # default is ""

  # second Kafka config entry
  - brokers: localhost:9092
    topics: 
      - test11
    group_id: k2ws-test-group2
    auto_offset: latest
    auto_commit: false
    addr: :8889
    ws_path: ws # default is ""
    test_path: "" # default is "test"
    include_headers: true # default is `false`
    message_type: json # default is "json"
```

This will start two web-socket servers, one on `ws://localhost:8888/onlyyouknow` and the other one on `ws://localhost:8889/ws`.
To test them in browser you can open `ws://localhost:8888/onlyyouknow/test` and `ws://localhost:8889/`.

To serve *HTTPS* just set path to certificate file in `cert_file` and private key file in `key_file` ( see comments in the example above ).

##### Kafka config entry options

Property           |Required | Range           |       Default | Description              
-------------------|:-------:|-----------------|--------------:|--------------------------
`brokers`          |   yes   |                 |               | Initial list of brokers as a CSV list of broker host or host:port.
`topics`           |         |                 |               | List of Kafka topics that will be served via websocket. If omitted topic list will be expected to be passed by client.
`group_id`         |         |                 |               | Client group id string. All clients sharing the same group.id belong to the same group. If omitted group id will be expected to be passed by client.
`auto_offset`      |         | smallest, earliest, beginning, largest, latest, end, error | largest | Action to take when there is no initial offset in offset store or the desired offset is out of range: 'smallest','earliest' - automatically reset the offset to the smallest offset, 'largest','latest' - automatically reset the offset to the largest offset, 'error' - trigger an error which is retrieved by consuming messages and checking. If omitted auto offset will be expected to be passed by client.
`auto_commit`      |         | `true`, `false` |       `false` | Automatically and periodically commit offsets in the background.
`addr`             |   yes   |                 |               | Host (or IP) and port pair where socket will be served from. Host (IP) is optional. Example `localhost:8888` or `:8888`.secret
`secret`           |         |                 |               | Additional route to websocket and test path. By default is empty.
`ws_path`          |         |                 |               | Path to websocket URL. By default it's empty.
`test_path`        |         |                 |          test | Path to test page URL.
`include_headers`  |         |                 |       `false` | Include headers into websocket message payload. Message will be in JSON format.
`message_type`     |         |   json, text    |          json | Type of Kafka messages. This is only important when `include_headers` option is set to `true` because it will affect creation of websocket message payload.

When `topics`, `group_id` and/or `auto_offset` are omitted in configuration, they are expected to be set by client as a query parameters in websocket URL. For example if websocket URL is `ws://localhost:8888/` client can set these like this `ws://localhost:8888/?topics=topicA,topicB&group_id=mygroup&auto_offset=earliest`. Note that this only works for parameters that are omitted from configuration thus setting them otherwise will have no effect.

You can serve more then one Kafka config entry on the same port as long as they all have unique websocket and test endpoints.

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
