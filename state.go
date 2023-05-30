package main

import (
	"io/ioutil"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
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

func (s *state) ToH() gin.H {
	return gin.H{
		"timer":  s.Timer.Format(time.RFC3339),
		"small":  s.Small,
		"big":    s.Big,
		"paused": !s.PauseTime.IsZero(),
	}
}

func readState() (state, error) {
	var s state
	contents, err := ioutil.ReadFile(conf.StateFile)
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
	return ioutil.WriteFile(conf.StateFile, contents, 0600)
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
