package main

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

// Config YAML config file
type Config struct {
	SchemaVersion string `yaml:"schema_version"`
	K2WS          []K2WS `yaml:"k2ws"`
}

// ReadK2WS read config file and returns collection of K2WS
func ReadK2WS(filename string) []K2WS {
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
	return config.K2WS
}
