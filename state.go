package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/akerl/go-lambda/s3"
	"gopkg.in/yaml.v2"
)

type state struct {
	Timer     time.Time `json:"timer"`
	Interval  int       `json:"interval"`
	Small     int       `json:"small"`
	Big       int       `json:"big"`
	PauseTime time.Time `json:"pausetime"`
}

type stateUpdate struct {
	Interval int  `form:"interval" json:"interval"`
	Small    int  `form:"small" json:"small"`
	Big      int  `form:"big" json:"big"`
	Pause    bool `form:"pause" json:"pause"`
	Resume   bool `form:"resume" json:"resume"`
}

type serializedState struct {
	Timer  string `json:"timer"`
	Small  int    `json:"small"`
	Big    int    `json:"big"`
	Paused bool   `json:"paused"`
}

func (s *state) ToJSON() ([]byte, error) {
	return json.Marshal(serializedState{
		Timer:  s.Timer.Format(time.RFC3339),
		Small:  s.Small,
		Big:    s.Big,
		Paused: !s.PauseTime.IsZero(),
	})
}

func readState() (state, error) {
	var s state
	contents, err := s3.GetObject(c.StateBucket, c.StateKey)
	if err != nil {
		return s, err
	}

	err = yaml.Unmarshal(contents, &s)
	return s, err
}

func writeState(s state) error {
	contents, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return s3.PutObject(c.StateBucket, c.StateKey, string(contents))
}

func checkState() (state, error) {
	s, err := readState()
	if err != nil {
		return s, err
	}

	if !s.PauseTime.IsZero() {
		delta := s.Timer.Sub(s.PauseTime)
		s.Timer = time.Now().Add(delta)
	} else if s.Timer.Before(time.Now()) {
		s.Timer = time.Now().Add(time.Minute * time.Duration(s.Interval))
		s.Small = s.Small * 2
		s.Big = s.Big * 2
		err = writeState(s)
		if err != nil {
			return s, err
		}
	}

	return s, nil
}

func updateState(su stateUpdate) error { //revive:disable-line:cyclomatic
	s, err := readState()
	if err != nil {
		return fmt.Errorf("failed to read state")
	}

	if su.Pause && s.PauseTime.IsZero() {
		s.PauseTime = time.Now()
	} else if su.Resume && !s.PauseTime.IsZero() {
		delta := s.Timer.Sub(s.PauseTime)
		s.Timer = time.Now().Add(delta)
		s.PauseTime = time.Time{}
	} else if su.Pause || su.Resume {
		return nil
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
	return writeState(s)
}
