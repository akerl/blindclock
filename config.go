package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
)

type config struct {
	StateFile string `json:"state_file"`
	SqsQueue  string `json:"sqs_queue"`
}

var conf config

func loadConfig() error {
	if len(os.Args) == 1 {
		return fmt.Errorf("must provide config file path")
	}
	configFile := os.Args[1]

	contents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(contents, &conf)
}
