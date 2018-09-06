package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/websocket"
)

// K2WS Kafka to websocket config
type K2WS struct {
	Address     string
	TLSCertFile string
	TLSKeyFile  string
	WebSockets  map[string]*K2WSKafka
	TestUIs     map[string]*string
}

// K2WSKafka Kafka config
type K2WSKafka struct {
	KafkaConsumerConfig     kafka.ConfigMap
	KafkaDefaultTopicConfig kafka.ConfigMap
	KafkaTopics             []string
	IncludeHeaders          bool
	MessageType             string
}

var rexStatic = regexp.MustCompile(`(.*)(/static/.+(\.[a-z0-9]+))$`)
var mimeTypes = map[string]string{
	".js":   "application/javascript",
	".htm":  "text/html; charset=utf-8",
	".html": "text/html; charset=utf-8",
	".css":  "text/css; charset=utf-8",
	".json": "application/json",
	".xml":  "text/xml; charset=utf-8",
	".jpg":  "image/jpeg",
	".png":  "image/png",
	".svg":  "image/svg+xml",
	".gif":  "image/gif",
	".pdf":  "application/pdf",
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
	if k2ws.TLSCertFile != "" {
		return http.ListenAndServeTLS(k2ws.Address, k2ws.TLSCertFile, k2ws.TLSKeyFile, k2ws)
	}
	return http.ListenAndServe(k2ws.Address, k2ws)
}

func copyConfigMap(m kafka.ConfigMap) kafka.ConfigMap {
	nm := make(kafka.ConfigMap)
	for k, v := range m {
		nm[k] = v
	}
	return nm
}

func (k2ws *K2WS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	submatch := rexStatic.FindStringSubmatch(r.URL.Path)
	if len(submatch) > 0 {
		// serve static files
		testUIPath := submatch[1]
		if testUIPath == "" {
			testUIPath = "/"
		}
		if _, exists := k2ws.TestUIs[testUIPath]; exists {
			if payload, err := FSByte(false, submatch[2]); err == nil {
				var mime = "application/octet-stream"
				if m, ok := mimeTypes[submatch[3]]; ok {
					mime = m
				}
				w.Header().Add("Content-Type", mime)
				w.Header().Add("Content-Length", fmt.Sprintf("%d", len(payload)))
				w.Write(payload)
			}
		}
	} else if wsPath, exists := k2ws.TestUIs[r.URL.Path]; exists {
		// readTemplate()
		if k2ws.TLSCertFile != "" {
			homeTemplate.Execute(w, templateInfo{strings.TrimRight(r.URL.Path, "/"), "wss://" + r.Host + *wsPath})
		} else {
			homeTemplate.Execute(w, templateInfo{strings.TrimRight(r.URL.Path, "/"), "ws://" + r.Host + *wsPath})
		}
	} else if kcfg, exists := k2ws.WebSockets[r.URL.Path]; exists {
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
		config := copyConfigMap(kcfg.KafkaConsumerConfig)
		// config["debug"] = "protocol"
		// config["broker.version.fallback"] = "0.8.0"
		config["session.timeout.ms"] = 6000
		config["go.events.channel.enable"] = true
		config["go.application.rebalance.enable"] = true
		if config["group.id"] == nil {
			config["group.id"] = parseQueryString(query, "group.id", "")
		}
		defTopic := copyConfigMap(kcfg.KafkaDefaultTopicConfig)
		if defTopic["auto.offset.reset"] == nil {
			defTopic["auto.offset.reset"] = parseQueryString(query, "auto.offset.reset", "")
		}
		config["default.topic.config"] = defTopic
		topics := kcfg.KafkaTopics
		if len(topics) == 0 {
			if t := parseQueryString(query, "topics", ""); t != "" {
				topics = strings.Split(t, ",")
			} else {
				log.Printf("No topic(s)")
				return
			}
		}

		// Instantiate consumer
		consumer, err := kafka.NewConsumer(&config)
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
					if kcfg.IncludeHeaders {
						headers := ""
						for _, header := range e.Headers {
							headers += fmt.Sprintf(",\"%s\":\"%s\"", header.Key, string(header.Value))
						}
						if len(headers) > 0 {
							headers = headers[1:]
						}
						var msg string
						if kcfg.MessageType == "json" {
							msg = fmt.Sprintf("{\"headers\":{%s},\"value\":%s}", headers, string(e.Value))
						} else {
							val, err := json.Marshal(string(e.Value))
							if err == nil {
								msg = fmt.Sprintf("{\"headers\":{%s},\"value\":%s}", headers, string(val))
							} else {
								msg = fmt.Sprintf("{\"headers\":{%s}}", headers)
							}
						}
						err = wscon.WriteMessage(websocket.TextMessage, []byte(msg))
					} else {
						err = wscon.WriteMessage(websocket.TextMessage, e.Value)
					}
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
	}
	w.WriteHeader(404)
}
