package main

import (
	"fmt"

	"github.com/akerl/go-lambda/s3"
)

type config struct {
	SlackTokens []string `json:"slacktokens"`
	StateBucket string   `json:"statebucket"`
	StateKey    string   `json:"statekey"`
}

func loadConfig() error {
	cf, err := s3.GetConfigFromEnv(&c)
	if err != nil {
		return err
	}

	cf.OnError = func(_ *s3.ConfigFile, err error) {
		fmt.Println(err)
	}
	cf.Autoreload(60)

	return nil
}
