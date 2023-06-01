package main

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/akerl/go-lambda/apigw/events"
	"github.com/slack-go/slack"
)

var slackUpdateRegex = regexp.MustCompile(`^(\d+)(?: (\d+))?(?: (\d+))?$`)

func validSlackUser(userID string) bool {
	for _, i := range c.SlackUsers {
		if i == userID {
			return true
		}
	}
	return false
}

func buildSlackMessage(text string) (*slack.Msg, error) {
	return &slack.Msg{Text: text}, nil
}

func slackUpdate(req events.Request) (*slack.Msg, error) { //revive:disable-line:cyclomatic
	bodyParams, err := req.BodyAsParams()
	if err != nil {
		return buildSlackMessage("failed to parse body params")
	}

	if !validSlackUser(bodyParams["user_id"]) {
		return buildSlackMessage("unauthorized user")
	}

	var su stateUpdate
	text := bodyParams["text"]
	var msg string
	switch {
	case text == "pause":
		su.Pause = true
		msg = "Paused!"
	case text == "resume":
		su.Resume = true
		msg = "Resumed!"
	case text == "toggle":
		su.Resume = true
		su.Pause = true
		msg = "Toggled!"
	case slackUpdateRegex.MatchString(text):
		match := slackUpdateRegex.FindStringSubmatch(text)
		su.Small, _ = strconv.Atoi(match[1])
		if match[2] == "" {
			su.Big = su.Small * 2
		} else {
			su.Big, _ = strconv.Atoi(match[2])
		}
		if match[3] != "" {
			su.Interval, _ = strconv.Atoi(match[3])
		} else {
			su.Interval = -1
		}
		msg = fmt.Sprintf("Setting blinds to %d / %d", su.Small, su.Big)
	default:
		return buildSlackMessage("invalid input")
	}

	if err := updateState(su); err != nil {
		return buildSlackMessage("state update failed")
	}
	return buildSlackMessage(msg)
}
