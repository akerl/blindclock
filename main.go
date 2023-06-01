package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/akerl/go-lambda/mux"
	"github.com/akerl/go-lambda/mux/receivers/slack"
)

var (
	c *config

	stateRegex   = regexp.MustCompile(`^/state$`)
	defaultRegex = regexp.MustCompile(`^/.*$`)
)

func main() {
	if err := loadConfig(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	d := mux.NewDispatcher(
		&slack.Handler{
			HandleFunc:    slackUpdate,
			SigningTokens: c.SlackTokens,
		},
		mux.NewRoute(stateRegex, stateHandler),
		mux.NewRoute(defaultRegex, defaultHandler),
	)
	mux.Start(d)
}
