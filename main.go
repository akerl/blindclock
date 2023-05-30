package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/akerl/go-lambda/mux"
)

var (
	c *config

	indexRegex   = regexp.MustCompile(`^/$`)
	faviconRegex = regexp.MustCompile(`^/favicon.ico$`)
	fontRegex    = regexp.MustCompile(`^/font.ttf$`)
	stateRegex   = regexp.MustCompile(`^/state$`)
	defaultRegex = regexp.MustCompile(`^/.*$`)
)

func main() {
	if err := loadConfig(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	d := mux.NewDispatcher(
		mux.NewRoute(indexRegex, indexHandler),
		mux.NewRoute(faviconRegex, faviconHandler),
		mux.NewRoute(fontRegex, fontHandler),
		mux.NewRoute(stateRegex, stateHandler),
		mux.NewRoute(defaultRegex, defaultHandler),
	)
	mux.Start(d)
}
