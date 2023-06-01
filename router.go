package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/akerl/go-lambda/apigw/events"
	"github.com/slack-go/slack"
)

//go:embed assets/favicon.ico assets/index.html assets/fonts/Roboto-Thin.ttf
var static embed.FS

var slackUpdateRegex = regexp.MustCompile(`^(\d+)(?: (\d+))?(?: (\d+)?$`)

func defaultHandler(req events.Request) (events.Response, error) {
	return events.Redirect("https://"+req.Headers["Host"], 303)
}

func missingHandler(_ events.Request) (events.Response, error) {
	return events.Respond(404, "resource does not exist")
}

func indexHandler(req events.Request) (events.Response, error) {
	content, err := static.ReadFile("assets/index.html")
	if err != nil {
		return missingHandler(req)
	}
	return events.Response{
		StatusCode: 200,
		Body:       string(content),
		Headers:    map[string]string{"Content-Type": "text/html; charset=utf-8"},
	}, nil
}

func faviconHandler(req events.Request) (events.Response, error) {
	content, err := static.ReadFile("assets/favicon.ico")
	if err != nil {
		return missingHandler(req)
	}
	return events.Response{
		StatusCode:      200,
		Body:            base64.StdEncoding.EncodeToString(content),
		Headers:         map[string]string{"Content-Type": "image/x-icon"},
		IsBase64Encoded: true,
	}, nil
}

func fontHandler(req events.Request) (events.Response, error) {
	content, err := static.ReadFile("assets/fonts/Roboto-Thin.ttf")
	if err != nil {
		return missingHandler(req)
	}
	return events.Response{
		StatusCode:      200,
		Body:            base64.StdEncoding.EncodeToString(content),
		Headers:         map[string]string{"Content-Type": "font/ttf"},
		IsBase64Encoded: true,
	}, nil
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

func parseSlackPost(req events.Request) (stateUpdate, error) {
	var su stateUpdate
	bodyParams, err := req.BodyAsParams()
	if err != nil {
		return su, err
	}
	text := bodyParams["text"]
	switch {
	case text == "pause":
		su.Pause = true
	case text == "resume":
		su.Resume = true
	case text == "toggle":
		su.Resume = true
		su.Pause = true
	case slackUpdateRegex.MatchString(text):
		match := slackUpdateRegex.FindStringSubmatch(text)
		switch len(match) {
		case 1:
			su.Small, _ = strconv.Atoi(match[0])
			su.Big = su.Small * 2
		case 2:
			su.Small, _ = strconv.Atoi(match[0])
			su.Big, _ = strconv.Atoi(match[1])
		case 3:
			su.Small, _ = strconv.Atoi(match[0])
			su.Big, _ = strconv.Atoi(match[1])
			su.Interval, _ = strconv.Atoi(match[2])
		}
	default:
		return su, fmt.Errorf("invalid input")
	}

	return su, nil
}

func statePost(req events.Request) (events.Response, error) { //revive:disable-line:cyclomatic
	if resp, err := slackAuth(req); err != nil {
		return resp, err
	}

	su, err := parseSlackPost(req)
	if err != nil {
		return events.Fail("failed to parse")
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
