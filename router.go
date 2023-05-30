package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func writeOrError(s state, c *gin.Context) {
	err := writeState(s)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, s.ToH())
	}
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

	s, err := readState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

	writeOrError(s, c)
}

func pause(c *gin.Context) {
	s, err := readState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if s.PauseTime.IsZero() {
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
	if !s.PauseTime.IsZero() {
		delta := s.Timer.Sub(s.PauseTime)
		s.Timer = time.Now().Add(delta)
		s.PauseTime = time.Time{}
	}
	writeOrError(s, c)
}
