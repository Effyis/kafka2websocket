package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// ConfigK2WS Kafka to websocket YAML
type ConfigK2WS struct {
	Brokers        string   `yaml:"brokers"`
	Topics         []string `yaml:"topics"`
	GroupID        string   `yaml:"group_id"`
	AutoOffset     string   `yaml:"auto_offset"`
	AutoCommit     bool     `yaml:"auto_commit"`
	Addr           string   `yaml:"addr"`
	Secret         string   `yaml:"secret"`
	TestPath       string   `yaml:"test_path"`
	WSPath         string   `yaml:"ws_path"`
	IncludeHeaders bool     `yaml:"include_headers"`
	MessageType    string   `yaml:"message_type"`
}

// Config YAML config file
type Config struct {
	SchemaVersion string       `yaml:"schema_version"`
	ConfigK2WSs   []ConfigK2WS `yaml:"k2ws"`
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
	k2wsMap := make(map[string]*K2WS)
	for _, kwsc := range config.ConfigK2WSs {
		var k2ws *K2WS
		var exists bool
		if k2ws, exists = k2wsMap[kwsc.Addr]; !exists {
			k2ws = &K2WS{
				Addr: kwsc.Addr,
				WS:   make(map[string]*K2WSKafka),
				Test: make(map[string]*string),
			}
			k2wsMap[kwsc.Addr] = k2ws
		}
		if kwsc.MessageType == "" {
			kwsc.MessageType = "json"
		}
		testPath := kwsc.TestPath
		wsPath := kwsc.WSPath
		if testPath == "" && wsPath == "" {
			testPath = "test"
		}
		if kwsc.Secret != "" {
			testPath = kwsc.Secret + "/" + testPath
			wsPath = kwsc.Secret + "/" + wsPath
		}
		testPath = "/" + strings.TrimRight(testPath, "/")
		wsPath = "/" + strings.TrimRight(wsPath, "/")

		if testPath == wsPath {
			panic(fmt.Sprintf("test_path and ws_path can't be same [%s]", kwsc.TestPath))
		}
		if kwsc.Brokers == "" {
			panic(fmt.Sprintf("brokers must be defined, address [%s]", kwsc.Addr))
		}
		if _, exists := k2ws.Test[testPath]; exists {
			panic(fmt.Sprintf("test path [%s] already defined", testPath))
		}
		if _, exists := k2ws.WS[testPath]; exists {
			panic(fmt.Sprintf("test path [%s] already defined as websocket path", testPath))
		}
		if _, exists := k2ws.WS[wsPath]; exists {
			panic(fmt.Sprintf("websocket path [%s] already defined", wsPath))
		}
		if _, exists := k2ws.Test[wsPath]; exists {
			panic(fmt.Sprintf("websocket path [%s] already defined as test path", wsPath))
		}
		if kwsc.MessageType != "json" && kwsc.MessageType != "text" {
			panic(fmt.Sprintf("invalid message_type [%s]", kwsc.MessageType))
		}
		k2ws.Test[testPath] = &wsPath
		k2ws.WS[wsPath] = &K2WSKafka{
			Brokers:        kwsc.Brokers,
			Topics:         kwsc.Topics,
			GroupID:        kwsc.GroupID,
			AutoOffset:     kwsc.AutoOffset,
			AutoCommit:     kwsc.AutoCommit,
			IncludeHeaders: kwsc.IncludeHeaders,
			MessageType:    kwsc.MessageType,
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
