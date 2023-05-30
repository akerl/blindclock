package main

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

const stateFile = "state.yml"

type state struct {
	Timer     time.Time `json:"timer"`
	Interval  int       `json:"interval"`
	Small     int       `json:"small"`
	Big       int       `json:"big"`
	Pause     bool      `json:"pause"`
	PauseTime time.Time `json:"pausetime"`
}

type stateUpdate struct {
	Interval int `form:"interval"`
	Small    int `form:"small"`
	Big      int `form:"big"`
}

func (s *state) ToH() gin.H {
	return gin.H{
		"timer": s.Timer.Format(time.RFC3339),
		"small": s.Small,
		"big":   s.Big,
		"pause": s.Pause}
}

func readState() (state, error) {
	var s state
	contents, err := ioutil.ReadFile(stateFile)
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
	return ioutil.WriteFile(stateFile, contents, 0600)
}

func writeOrError(s state, c *gin.Context) {
	err := writeState(s)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, s.ToH())
	}
}

func checkState() (state, error) {
	s, err := readState()
	if err != nil {
		return s, err
	}

	if s.Pause {
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

func getState(c *gin.Context) { //revive:disable-line:get-return
	s, err := checkState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, s.ToH())
	}
}

func postState(c *gin.Context) {
	var su stateUpdate
	if err := c.ShouldBind(&su); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s := state{
		Timer:    time.Now().Add(time.Minute * time.Duration(su.Interval)),
		Interval: su.Interval,
		Small:    su.Small,
		Big:      su.Big,
	}
	writeOrError(s, c)
}

func pause(c *gin.Context) {
	s, err := readState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !s.Pause {
		s.Pause = true
		s.PauseTime = time.Now()
	}
	writeOrError(s, c)
}

func resume(c *gin.Context) {
	s, err := readState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if s.Pause && !s.PauseTime.IsZero() {
		s.Pause = false
		delta := s.Timer.Sub(s.PauseTime)
		s.Timer = time.Now().Add(delta)
	}
	s.Pause = false
	writeOrError(s, c)
}

func main() {
	router := gin.Default()
	router.StaticFile("/favicon.ico", "./public/favicon.ico")
	router.StaticFile("/", "./public/index.html")
	router.GET("/state", getState)
	router.POST("/state", postState)
	router.POST("/pause", pause)
	router.POST("/resume", resume)
	router.Run(":8080")
}
