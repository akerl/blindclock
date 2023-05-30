//go:generate resources -output static.go -declare -var static -fmt -trim assets/ assets/*
package main

import (
	"encoding/json"
	"time"

	"github.com/akerl/go-lambda/apigw/events"
	"github.com/slack-go/slack"
)

func defaultHandler(req events.Request) (events.Response, error) {
	return events.Redirect("https://"+req.Headers["Host"], 303)
}

func missingHandler(_ events.Request) (events.Response, error) {
	return events.Respond(404, "resource does not exist")
}

func staticLookup(req events.Request, path, contentType string) (events.Response, error) {
	content, found := static.String(path)
	if !found {
		return missingHandler(req)
	}
	return events.Response{
		StatusCode: 200,
		Body:       content,
		Headers: map[string]string{
			"Content-Type": contentType,
		},
	}, nil
}

func indexHandler(req events.Request) (events.Response, error) {
	return staticLookup(req, "/index.html", "text/html; charset=utf-8")
}

func faviconHandler(req events.Request) (events.Response, error) {
	return staticLookup(req, "/favicon.ico", "image/x-icon")
}

func fontHandler(req events.Request) (events.Response, error) {
	return staticLookup(req, "/fonts/Roboto-Thin.ttf", "font/ttf")
}

func stateHandler(req events.Request) (events.Response, error) {
	switch req.HTTPMethod {
	case "GET":
		return stateGet(req)
	case "POST":
		return statePost(req)
	default:
		return events.Respond(405, "unsupported method")
	}
}

func stateGet(_ events.Request) (events.Response, error) {
	s, err := checkState()
	if err != nil {
		return events.Fail(err.Error())
	}
	body, err := json.Marshal(s)
	if err != nil {
		return events.Fail(err.Error())
	}
	return events.Response{
		StatusCode: 200,
		Body:       string(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func slackAuth(req events.Request) (events.Response, error) {
	if len(c.SlackTokens) == 0 {
		return events.Reject("no signing tokens provided")
	}

	byteBody := []byte(req.Body)

	for _, i := range c.SlackTokens {
		sv, err := slack.NewSecretsVerifier(req.MultiValueHeaders, i)
		if err != nil {
			return events.Reject("failed to create secret verifier")
		}
		if _, err := sv.Write(byteBody); err != nil {
			return events.Reject("failed to parse body")
		}
		if err := sv.Ensure(); err == nil {
			return events.Response{}, nil
		}
	}

	return events.Reject("invalid signature")
}

func statePost(req events.Request) (events.Response, error) { //revive:disable-line:cyclomatic
	if resp, err := slackAuth(req); err != nil {
		return resp, err
	}

	var su stateUpdate
	if err := json.Unmarshal([]byte(req.Body), &su); err != nil {
		return events.Fail("failed to unmarshal")
	}

	s, err := readState()
	if err != nil {
		return events.Fail("failed to read state")
	}

	if su.Pause && s.PauseTime.IsZero() {
		s.PauseTime = time.Now()
	} else if su.Resume && !s.PauseTime.IsZero() {
		delta := s.Timer.Sub(s.PauseTime)
		s.Timer = time.Now().Add(delta)
		s.PauseTime = time.Time{}
	} else if su.Pause || su.Resume {
		return events.Succeed("")
	} else {
		if su.Interval != 0 {
			s.Timer = time.Now().Add(time.Minute * time.Duration(su.Interval))
			s.Interval = su.Interval
			s.PauseTime = time.Time{}
		}
		if su.Small != 0 {
			s.Small = su.Small
		}
		if su.Big != 0 {
			s.Big = su.Big
		}
	}
	if err := writeState(s); err != nil {
		return events.Fail("failed to write state")
	}
	return events.Succeed("")
}
