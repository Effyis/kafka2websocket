package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/websocket"
)

// K2WS Kafka to websocket config
type K2WS struct {
	Addr string
	WS   map[string]*K2WSKafka
	Test map[string]*string
}

// K2WSKafka Kafka config
type K2WSKafka struct {
	Brokers    string
	Topics     []string
	GroupID    string
	AutoOffset string
	AutoCommit bool
}

func parseQueryString(query url.Values, key string, val string) string {
	if val == "" {
		if vals, exists := query[key]; exists && len(vals) > 0 {
			return vals[0]
		}
	}
	return val
}

// Start start websocket and start consuming from Kafka topic(s)
func (k2ws *K2WS) Start() error {
	return http.ListenAndServe(k2ws.Addr, k2ws)
}

func (k2ws *K2WS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if wsPath, exists := k2ws.Test[r.URL.Path]; exists {
		// readTemplate()
		homeTemplate.Execute(w, "ws://"+r.Host+*wsPath)
	} else if kcfg, exists := k2ws.WS[r.URL.Path]; exists {
		// Upgrade to websocket connection
		upgrader := websocket.Upgrader{}
		wscon, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Websocket http upgrade failed: %v\n", err)
			return
		}
		defer wscon.Close()

		// Read kafka params from query string
		query := r.URL.Query()
		groupID := parseQueryString(query, "group_id", kcfg.GroupID)
		autoOffset := parseQueryString(query, "auto_offset", kcfg.AutoOffset)
		topics := kcfg.Topics
		if len(topics) == 0 {
			if t := parseQueryString(query, "topics", ""); t != "" {
				topics = strings.Split(t, ",")
			} else {
				log.Printf("No topic(s)")
				return
			}
		}

		// Instantiate consumer
		consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":               kcfg.Brokers,
			"group.id":                        groupID,
			"default.topic.config":            kafka.ConfigMap{"auto.offset.reset": autoOffset},
			"auto.commit.enable":              kcfg.AutoCommit,
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
		for _, topic := range topics {
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

		// Subscribe consumer to the topics
		err = consumer.SubscribeTopics(topics, nil)
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
