package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/websocket"
)

// K2WS Kafka to websocket YAML
type K2WS struct {
	Brokers    string   `yaml:"brokers"`
	Topics     []string `yaml:"topics"`
	GroupID    string   `yaml:"group_id"`
	AutoOffset string   `yaml:"auto_offset"`
	AutoCommit bool     `yaml:"auto_commit"`
	Addr       string   `yaml:"addr"`
	Secret     string   `yaml:"secret"`
}

// Start start websocket and start consuming from Kafka topic(s)
func (kws *K2WS) Start() error {
	return http.ListenAndServe(kws.Addr, kws)
}

func (kws *K2WS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == kws.patternTest() {
		// readTemplate()
		homeTemplate.Execute(w, "ws://"+r.Host+kws.patternWS())
	} else if r.URL.Path == kws.patternWS() {
		// Upgrade to websocket connection
		upgrader := websocket.Upgrader{}
		wscon, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Websocket http upgrade failed: %v\n", err)
			return
		}
		defer wscon.Close()

		// Instantiate consumer
		consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":               kws.Brokers,
			"group.id":                        kws.GroupID,
			"default.topic.config":            kafka.ConfigMap{"auto.offset.reset": kws.AutoOffset},
			"auto.commit.enable":              kws.AutoCommit,
			"session.timeout.ms":              6000,
			"go.events.channel.enable":        true,
			"go.application.rebalance.enable": true,
			// "broker.version.fallback":         "0.8.0",
			// "debug":                           "protocol",
		})
		if err != nil {
			fmt.Printf("Can't create consumer: %v\n", err)
			return
		}
		defer consumer.Close()

		// Make sure all topics actually exist
		meta, err := consumer.GetMetadata(nil, true, 5000)
		if err != nil {
			fmt.Printf("Can't get metadata: %v\n", err)
			return
		}
		for _, topic := range kws.Topics {
			if _, exists := meta.Topics[topic]; !exists {
				log.Printf("Topic [%s] doesn't exist", topic)
				return
			}
		}

		// Make sure to read client message and react on close/error
		chClose := make(chan bool)
		go func() {
			for {
				_, _, err := wscon.ReadMessage()
				if err != nil {
					chClose <- true
					if !strings.HasPrefix(err.Error(), "websocket: close") {
						log.Printf("WebSocket read error: %v\n", err)
					}
					return
				}
			}
		}()

		// Subscribe costumer to the topics
		err = consumer.SubscribeTopics(kws.Topics, nil)
		if err != nil {
			fmt.Printf("Can't subscribe costumer: %v\n", err)
			return
		}
		defer consumer.Unsubscribe()

		log.Printf("Websocket opened %s\n", r.Host)
		running := true
		// Keep reading and sending messages
		for running {
			select {
			// Exit if websocket read fails
			case <-chClose:
				running = false
			case ev := <-consumer.Events():
				switch e := ev.(type) {
				case kafka.AssignedPartitions:
					consumer.Assign(e.Partitions)
				case kafka.RevokedPartitions:
					consumer.Unassign()
				case *kafka.Message:
					err = wscon.WriteMessage(websocket.TextMessage, e.Value)
					if err != nil {
						log.Printf("WebSocket write error: %v (%v)\n", err, e)
						running = false
					}
				case kafka.PartitionEOF:
					// log.Printf("%% Reached %v\n", e)
				case kafka.Error:
					log.Printf("%% Error: %v\n", e)
					running = false
				}
			}
		}
		log.Printf("Websocket closed %s\n", r.Host)
	} else {
		w.WriteHeader(404)
	}
}

func (kws *K2WS) patternTest() string {
	if kws.Secret != "" {
		return fmt.Sprintf("/%s/test", kws.Secret)
	}
	return "/test"
}

func (kws *K2WS) patternWS() string {
	if kws.Secret != "" {
		return fmt.Sprintf("/%s", kws.Secret)
	}
	return "/"
}
