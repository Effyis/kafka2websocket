package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type jsonTopicPartition struct {
	Topic     *string `json:"topic"`
	Partition int32   `json:"partition"`
	Offset    int64   `json:"offset"`
	Error     *error  `json:"error"`
}

type jsonMessage struct {
	Key            string              `json:"key"`
	Headers        []jsonHeader        `json:"headers"`
	TopicPartition jsonTopicPartition  `json:"partition"`
	Timestamp      time.Time           `json:"timestamp"`
	TimestampType  kafka.TimestampType `json:"timestampType"`
}

type jsonHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var rexJSONVal = regexp.MustCompile(`}$`)

// JSONBytefy coverts kafka Message into JSON byte slice
func JSONBytefy(msg *kafka.Message, messageType string) ([]byte, error) {
	var jsonMsg = jsonMessage{
		Key:           string(msg.Key),
		Timestamp:     msg.Timestamp,
		TimestampType: msg.TimestampType,
	}
	jsonMsg.TopicPartition = jsonTopicPartition{
		Topic:     msg.TopicPartition.Topic,
		Partition: msg.TopicPartition.Partition,
		Offset:    int64(msg.TopicPartition.Offset),
		Error:     &msg.TopicPartition.Error,
	}

	for _, header := range msg.Headers {
		jsonMsg.Headers = append(jsonMsg.Headers, jsonHeader{Key: header.Key, Value: string(header.Value)})
	}

	b, err := json.Marshal(jsonMsg)
	var val string
	if messageType == "json" {
		val = ",\"value\":" + string(msg.Value) + "}"
	} else if messageType == "binary" {
		val = ",\"value\":\"" + base64.StdEncoding.EncodeToString(msg.Value) + "\"}"
	} else {
		if jsonVal, err := json.Marshal(string(msg.Value)); err == nil {
			val = ",\"value\":" + string(jsonVal) + "}"
		} else {
			err = fmt.Errorf("Can't stringify value as string: %v", err)
		}
	}

	return rexJSONVal.ReplaceAll(b, []byte(val)), err
}
