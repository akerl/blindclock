package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"

	"github.com/akerl/go-lambda/apigw/events"
)

//go:embed assets/favicon.ico assets/index.html assets/fonts/Roboto-Thin.ttf
//go:embed assets/images/money.svg assets/images/clock.svg assets/images/timer.svg
var static embed.FS

func defaultHandler(req events.Request) (events.Response, error) {
	return events.Redirect("https://"+req.Headers["Host"], 303)
}

func indexHandler(_ events.Request) (events.Response, error) {
	content, err := static.ReadFile("assets/index.html")
	if err != nil {
		return events.Fail("failed to load content")
	}
	return events.Response{
		StatusCode: 200,
		Body:       string(content),
		Headers:    map[string]string{"Content-Type": "text/html; charset=utf-8"},
	}, nil
}

func faviconHandler(_ events.Request) (events.Response, error) {
	content, err := static.ReadFile("assets/favicon.ico")
	if err != nil {
		return events.Fail("failed to load content")
	}
	return events.Response{
		StatusCode:      200,
		Body:            base64.StdEncoding.EncodeToString(content),
		Headers:         map[string]string{"Content-Type": "image/x-icon"},
		IsBase64Encoded: true,
	}, nil
}

func fontHandler(_ events.Request) (events.Response, error) {
	content, err := static.ReadFile("assets/fonts/Roboto-Thin.ttf")
	if err != nil {
		return events.Fail("failed to load content")
	}
	return events.Response{
		StatusCode:      200,
		Body:            base64.StdEncoding.EncodeToString(content),
		Headers:         map[string]string{"Content-Type": "font/ttf"},
		IsBase64Encoded: true,
	}, nil
}

func moneyHandler(_ events.Request) (events.Response, error) {
	content, err := static.ReadFile("assets/images/money.svg")
	if err != nil {
		return events.Fail("failed to load content")
	}
	return events.Response{
		StatusCode: 200,
		Body:       string(content),
		Headers:    map[string]string{"Content-Type": "image/svg+xml"},
	}, nil
}

func clockHandler(_ events.Request) (events.Response, error) {
	content, err := static.ReadFile("assets/images/clock.svg")
	if err != nil {
		return events.Fail("failed to load content")
	}
	return events.Response{
		StatusCode: 200,
		Body:       string(content),
		Headers:    map[string]string{"Content-Type": "image/svg+xml"},
	}, nil
}
func timerHandler(_ events.Request) (events.Response, error) {
	content, err := static.ReadFile("assets/images/timer.svg")
	if err != nil {
		return events.Fail("failed to load content")
	}
	return events.Response{
		StatusCode: 200,
		Body:       string(content),
		Headers:    map[string]string{"Content-Type": "image/svg+xml"},
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

func validAuthToken(token string) bool {
	for _, i := range c.AuthTokens {
		if i == token {
			return true
		}
	}
	return false
}

func statePost(req events.Request) (events.Response, error) {
	if !validAuthToken(req.Headers["X-API-Key"]) {
		return events.Fail("unauthorized")
	}

	var su stateUpdate
	if err := json.Unmarshal([]byte(req.Body), &su); err != nil {
		return events.Fail("failed to unmarshal")
	}

	if err := updateState(su); err != nil {
		return events.Fail("failed to write state")
	}
	return events.Succeed("")
}
