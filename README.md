**Apache Kafka to WebSocket** creates web-socket server that serves data from Apache Kafka topic(s). This is also great tool for inspecting Kafka topic(s) using built-in browser based consumer.

## How to use
* create your `config.yaml` file
* download and run [binary](https://github.com/Effyis/kafka2websocket/releases) for your OS (`config.yaml` should be in the same directory)

### Simple setup
![Simple K2WS](https://raw.githubusercontent.com/Effyis/kafka2websocket/master/docs/imgs/k2ws%20simple.png)

```yaml
schema.version: "1.0"
kafka.to.websocket:
  - kafka.consumer.config:
      metadata.broker.list: kafka:9092
      group.id: k2ws-group
    kafka.topics:
      - test
    address: :80
```

### Complex setup
![Complex K2WS](https://raw.githubusercontent.com/Effyis/kafka2websocket/master/docs/imgs/k2ws%20complex.png)
```yaml
schema.version: "1.0"
# domain certificate files are located in certs folder
tls.cert.file: ./certs/k2ws.socialgist.com.crt
tls.key.file: ./certs/k2ws.socialgist.com.key
kafka.to.websocket:
  # WebSocket X | wss://k2ws.socialgist.com/kafkaA
  - kafka.consumer.config:
      metadata.broker.list: kafkaA:9092
      group.id: k2ws-wsx-group
    kafka.topics:
      - a1
      - a2
    address: :443
    endpoint.prefix: kafka_a
  # WebSocket X | wss://k2ws.socialgist.com/kafkaB
  - kafka.consumer.config:
      metadata.broker.list: kafkaB:9092
      group.id: k2ws-wsx-group
    kafka.topics:
      - b1
    address: :443
    endpoint.prefix: kafka_b
  # WebSocket Y | wss://k2ws.socialgist.com/kafkaC
  - kafka.consumer.config:
      metadata.broker.list: kafkaC:9092
      group.id: k2ws-wsy-group
    kafka.topics:
      - c1
    address: :443
    endpoint.prefix: kafka_c
```

### Kafka config options
For setting up `kafka.consumer.config` and `kafka.default.topic.config ` refer to [librdkafka](https://github.com/edenhill/librdkafka/blob/dbde254bf7671d5b106ebc70de25ee92cd5fe6a7/CONFIGURATION.md) docs.

Property                                |Required | Range           |       Default | Description              
----------------------------------------|:-------:|-----------------|--------------:|--------------------------
`kafka.consumer.config/metadata.broker.list` |   yes   |                 |               | Initial list of brokers as a CSV list of broker host or host:port.
`kafka.consumer.config/group.id`             |         |                 |               | Client group id string. All clients sharing the same group.id belong to the same group. If omitted group id will be expected to be passed by client.
`kafka.default.topic.config/auto.offset.reset` |         | smallest, earliest, beginning, largest, latest, end, error |  | Action to take when there is no initial offset in offset store or the desired offset is out of range. If omitted it will be expected to be passed by client. 
`topics`                                |         |                 |               | List of Kafka topics that will be served via websocket. If omitted topic list will be expected to be passed by client.
`address`                               |   yes   |                 |               | Host (or IP) and port pair where socket will be served from. Host (IP) is optional. Example `localhost:8888` or `:8888`.
`endpoint.prefix`                       |         |                 |               | Prefix of the websocket and test paths. By default is empty.
`endpoint.websocket`                    |         |                 |               | Path to websocket URL. By default it's empty.
`endpoint.test`                         |         |                 |          test | Path to test page URL.
`message.details`                       |         |                 |       `false` | Include key, headers, topic partition, timesteamp into websocket message payload. Message will be in JSON format.
`message.type`                          |         |   json, text    |          json | Type of Kafka messages. This is only important when `message.details` option is set to `true` because it will affect creation of websocket message payload.

When `topics`, `kafka.consumer.config/group.id` and/or `kafka.consumer.config/auto.offset.reset` are omitted in configuration, they are expected to be set by client as a query parameters in websocket URL. For example if websocket URL is `ws://localhost:8888/` client can set these by making request to `ws://localhost:8888/?topics=topicA,topicB&group.id=mygroup&auto.offset.reset=earliest`. Note that this only works for parameters that are omitted from configuration thus setting them otherwise will have no effect.

You can serve more then one Kafka config entry on the same port as long as they all have unique websocket and test endpoints.

## Updating static HTTP files
Add/remove/modify files in `k2ws/static` directory and using [esc](https://github.com/mjibson/esc) tool from inside `k2ws` directory run `esc -o static.go static`. This will re-create `static.go`. 

## Build

**This project has [`librdkafka`](https://github.com/edenhill/librdkafka) dependency.**

You can build it statically with `go build -tags static`. Check out [confluentinc/confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go#static-builds) for more info.

### Build with Docker (Linux)
```sh
docker build -f Dockerfile.build -t k2ws-build .
docker run --rm -v $PWD/build:/build k2ws-build
```

You'll end up with `./build/k2ws` executable that works on Linux.

### Build and run with Docker
**Build**
```sh
docker build -f Dockerfile.k2ws -t k2ws .
```

**Run**
```sh
docker run -d --name=k2ws --net=host -v _PATH_TO_CONFIG_:/config.yaml k2ws
```

### Build for Windows
* get and start `cygwin64` installation from https://www.cygwin.com/setup-x86_64.exe
  * select install from Internet
  * select `C:\cygwin64` as root directory
  * select `muug.ca` domain
  * view -> full
  * install `x86_64-w64-mingw32-gcc` and `pkg-config`
    * click on the icon left from *Skip* ( keep clicking until you figure the latest version )
  * optionally put `setup-x86_64.exe` in `c:\cygwin64\` for convenience so that you can install more libs later if needed
* add `c:\cygwin64\` to PATH
* open your project directory in command prompt
* get `nuget` from https://dist.nuget.org/win-x86-commandline/latest/nuget.exe
* run `nuget install librdkafka.redist -Version 0.11.4` ( this is currently the latest version )
* this will download `rdkafka` into new `librdkafka.redist.0.11.4` directory
* copy `.\librdkafka.redist.0.11.4\build\native\include\` into `c:\cygwin64\usr\include\`
* copy `.\librdkafka.redist.0.11.4\build\native\lib\win7\x64\win7-x64-Release\v120\librdkafka.lib` into `c:\cygwin64\lib\librdkafka.a` (notice `.lib` is renamed to `.a`)
* create file `rdkafka.pc` in the project's root directory with following content:
```
prefix=c:/
libdir=c:/cygwin64/lib/
includedir=c:/cygwin64/usr/include

Name: librdkafka
Description: The Apache Kafka C/C++ library
Version: 0.11.4
Cflags: -I${includedir}
Libs: -L${libdir} -lrdkafka
Libs.private: -lssl -lcrypto -lcrypto -lz -ldl -lpthread -lrt
```
* run `set CC=x86_64-w64-mingw32-gcc` (this will allow `cgo` to use `x86_64-w64-mingw32-gcc` instead of `gcc` - you can make sure it worked with `go env CC`)
* run `go build`, that will create executable
* deliver `librdkafka.dll`, `msvcr120.dll` and `zlib.dll` from `.\librdkafka.redist.0.11.4\runtimes\win7-x64\native\` alongside with executable
