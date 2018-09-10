package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
)

func main() {
	configFile := flag.String("config", "config.yaml", "Config file location")
	flag.Parse()

	list := ReadK2WS(*configFile)
	for i := range list {
		go func(k2ws *K2WS) {
			err := k2ws.Start()
			if err != nil {
				log.Fatalln(err)
			}
		}(list[i])
	}

	var chExit = make(chan os.Signal, 1)
	signal.Notify(chExit, os.Interrupt)
	<-chExit
}
