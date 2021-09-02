package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	yaml "gopkg.in/yaml.v2"
)

const initConfig = `schema.version: "1.0"
# tls.cert.file: my-domain.crt
# tls.key.file: my-domain.key
kafka.to.websocket:
  - kafka.consumer.config:
      # https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md#global-configuration-properties
      metadata.broker.list: localhost:9092
      enable.auto.commit: false
      group.id: my-kafka-group
    kafka.default.topic.config:
      # https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md#topic-configuration-properties
      auto.offset.reset: latest
    kafka.topics:
      - my.kafka.topic
    address: :9999
    # message.details: false
    # message.type: json
    # endpoint.prefix: ""
    # endpoint.websocket: ws
    # endpoint.test: test
`

// ConfigK2WS Kafka to websocket YAML
type ConfigK2WS struct {
	KafkaConsumerConfig     kafka.ConfigMap `yaml:"kafka.consumer.config"`
	KafkaDefaultTopicConfig kafka.ConfigMap `yaml:"kafka.default.topic.config"`
	KafkaTopics             []string        `yaml:"kafka.topics"`
	Address                 string          `yaml:"address"`
	EndpointPrefix          string          `yaml:"endpoint.prefix"`
	EndpointTest            string          `yaml:"endpoint.test"`
	EndpointWS              string          `yaml:"endpoint.websocket"`
	MessageDetails          bool            `yaml:"message.details"`
	MessageType             string          `yaml:"message.type"`
	Compression             bool            `yaml:"compression"`
	CheckOrigin             string          `yaml:"checkorigin"`
}

// Config YAML config file
type Config struct {
	SchemaVersion string       `yaml:"schema.version"`
	TLSCertFile   string       `yaml:"tls.cert.file"`
	TLSKeyFile    string       `yaml:"tls.key.file"`
	ConfigK2WSs   []ConfigK2WS `yaml:"kafka.to.websocket"`
}

// ReadK2WS read config file and returns collection of K2WS
func ReadK2WS(filename string) []*K2WS {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error while reading config.yaml file: \n%v ", err)
	}
	log.Printf("%s\n%s", filename, string(fileContent))
	var config Config
	err = yaml.Unmarshal(fileContent, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	certFile := ""
	keyFile := ""
	if config.TLSCertFile != "" && config.TLSKeyFile == "" || config.TLSCertFile == "" && config.TLSKeyFile != "" {
		panic(fmt.Sprintf("Both certificate and key file must be defined"))
	} else if config.TLSCertFile != "" {
		if _, err := os.Stat(config.TLSCertFile); err == nil {
			if _, err := os.Stat(config.TLSKeyFile); err == nil {
				keyFile = config.TLSKeyFile
				certFile = config.TLSCertFile
			} else {
				panic(fmt.Sprintf("key file %s does not exist", config.TLSKeyFile))
			}
		} else {
			panic(fmt.Sprintf("certificate file %s does not exist", config.TLSKeyFile))
		}
	}
	k2wsMap := make(map[string]*K2WS)
	for _, kwsc := range config.ConfigK2WSs {
		var k2ws *K2WS
		var exists bool
		if k2ws, exists = k2wsMap[kwsc.Address]; !exists {
			k2ws = &K2WS{
				Address:     kwsc.Address,
				TLSCertFile: certFile,
				TLSKeyFile:  keyFile,
				WebSockets:  make(map[string]*K2WSKafka),
				TestUIs:     make(map[string]*string),
			}
			k2wsMap[kwsc.Address] = k2ws
		}
		if kwsc.MessageType == "" {
			kwsc.MessageType = "json"
		}
		testPath := kwsc.EndpointTest
		wsPath := kwsc.EndpointWS
		if testPath == "" && wsPath == "" {
			testPath = "test"
		}
		if kwsc.EndpointPrefix != "" {
			testPath = kwsc.EndpointPrefix + "/" + testPath
			wsPath = kwsc.EndpointPrefix + "/" + wsPath
		}
		testPath = "/" + strings.TrimRight(testPath, "/")
		wsPath = "/" + strings.TrimRight(wsPath, "/")

		if testPath == wsPath {
			panic(fmt.Sprintf("test path and websocket path can't be same [%s]", kwsc.EndpointTest))
		}
		if kwsc.KafkaConsumerConfig["metadata.broker.list"] == "" {
			panic(fmt.Sprintf("metadata.broker.list must be defined, address [%s]", kwsc.Address))
		}
		if kwsc.KafkaConsumerConfig["group.id"] == "" {
			panic(fmt.Sprintf("group.id must be defined, address [%s]", kwsc.Address))
		}
		if _, exists := k2ws.TestUIs[testPath]; exists {
			panic(fmt.Sprintf("test path [%s] already defined", testPath))
		}
		if _, exists := k2ws.WebSockets[testPath]; exists {
			panic(fmt.Sprintf("test path [%s] already defined as websocket path", testPath))
		}
		if _, exists := k2ws.WebSockets[wsPath]; exists {
			panic(fmt.Sprintf("websocket path [%s] already defined", wsPath))
		}
		if _, exists := k2ws.TestUIs[wsPath]; exists {
			panic(fmt.Sprintf("websocket path [%s] already defined as test path", wsPath))
		}
		if kwsc.MessageType != "json" &&
			kwsc.MessageType != "text" &&
			kwsc.MessageType != "binary" {
			panic(fmt.Sprintf("invalid message.type [%s]", kwsc.MessageType))
		}
        if kwsc.CheckOrigin == "" {
            kwsc.CheckOrigin = "hard"
        }
		if kwsc.CheckOrigin != "hard" &&
			kwsc.CheckOrigin != "danger" {
			panic(fmt.Sprintf("invalid checkorigin [%s]", kwsc.CheckOrigin))
		}
		k2ws.TestUIs[testPath] = &wsPath
		k2ws.WebSockets[wsPath] = &K2WSKafka{
			KafkaConsumerConfig:     kwsc.KafkaConsumerConfig,
			KafkaDefaultTopicConfig: kwsc.KafkaDefaultTopicConfig,
			KafkaTopics:             kwsc.KafkaTopics,
			MessageDetails:          kwsc.MessageDetails,
			MessageType:             kwsc.MessageType,
			Compression:             kwsc.Compression,
			CheckOrigin:             kwsc.CheckOrigin,
		}
	}
	k2wss := make([]*K2WS, len(k2wsMap))
	i := 0
	for _, k2ws := range k2wsMap {
		k2wss[i] = k2ws
		i++
	}
	return k2wss
}
