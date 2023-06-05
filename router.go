package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"

	"github.com/akerl/go-lambda/apigw/events"
)

//go:embed assets/favicon.ico assets/index.html assets/fonts/Roboto-Thin.ttf
var static embed.FS

type fileData struct {
	Path   string
	Type   string
	Binary bool
}

var fd = map[string]fileData{
	"/": fileData{
		Path: "assets/index.html",
		Type: "text/html; charset=utf-8",
	},
	"/favicon.ico": fileData{
		Path:   "assets/favicon.ico",
		Type:   "image/x-icon",
		Binary: true,
	},
	"/font.ttf": fileData{
		Path:   "assets/fonts/Roboto-Thin.ttf",
		Type:   "font/ttf",
		Binary: true,
	},
}

func defaultHandler(req events.Request) (events.Response, error) {
	data, ok := fd[req.Path]
	if !ok {
		return events.Redirect("https://"+req.Headers["Host"], 303)
	}

	content, err := static.ReadFile(data.Path)
	if err != nil {
		return events.Fail("failed to load content")
	}

	var body string
	if data.Binary {
		body = base64.StdEncoding.EncodeToString(content)
	} else {
		body = string(content)
	}

	return events.Response{
		StatusCode:      200,
		Body:            body,
		Headers:         map[string]string{"Content-Type": data.Type},
		IsBase64Encoded: data.Binary,
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
	body, err := s.ToJSON()
	if err != nil {
		return events.Fail(err.Error())
	}
	return events.Response{
		StatusCode: 200,
		Body:       string(body),
		Headers:    map[string]string{"Content-Type": "application/json"},
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
	if !validAuthToken(req.Headers["x-api-key"]) {
		return events.Fail("unauthorized")
	}

	body, err := req.DecodedBody()
	if err != nil {
		return events.Fail("failed to decode body")
	}

	var su stateUpdate
	if err := json.Unmarshal([]byte(body), &su); err != nil {
		return events.Fail("failed to unmarshal")
	}

	if err := updateState(su); err != nil {
		return events.Fail("failed to write state")
	}
	return events.Succeed("")
}
