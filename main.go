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

	indexRegex   = regexp.MustCompile(`^/$`)
	faviconRegex = regexp.MustCompile(`^/favicon.ico$`)
	fontRegex    = regexp.MustCompile(`^/font.ttf$`)
	moneyRegex   = regexp.MustCompile(`^/money.svg$`)
	clockRegex   = regexp.MustCompile(`^/clock.svg$`)
	timerRegex   = regexp.MustCompile(`^/timer.svg$`)
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
		mux.NewRoute(indexRegex, indexHandler),
		mux.NewRoute(faviconRegex, faviconHandler),
		mux.NewRoute(fontRegex, fontHandler),
		mux.NewRoute(moneyRegex, moneyHandler),
		mux.NewRoute(clockRegex, clockHandler),
		mux.NewRoute(timerRegex, timerHandler),
		mux.NewRoute(stateRegex, stateHandler),
		mux.NewRoute(defaultRegex, defaultHandler),
	)
	mux.Start(d)
}
