package main

import (
	"strconv"

	"github.com/akerl/go-lambda/apigw/events"
	"github.com/slack-go/slack"
)

func validSlackUser(userID string) bool {
	for _, i := range c.SlackUsers {
		if i == userID {
			return true
		}
	}
	return false
}

func buildSlackMessage(text string) (*slack.Msg, error) {
	return &slack.Msg{
		Text:         text,
		ResponseType: "in_channel",
	}, nil
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
		return buildSlackMessage("invalid input")
	}

	if err := updateState(su); err != nil {
		return buildSlackMessage("state update failed")
	}
	return buildSlackMessage("")
}
