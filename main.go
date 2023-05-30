package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
)

type config struct {
	StateFile string `json:"state_file"`
	SqsQueue  string `json:"sqs_queue"`
}

type state struct {
	Timer     time.Time `json:"timer"`
	Interval  int       `json:"interval"`
	Small     int       `json:"small"`
	Big       int       `json:"big"`
	Pause     bool      `json:"pause"`
	PauseTime time.Time `json:"pausetime"`
}

type stateUpdate struct {
	Interval int `form:"interval" json:"interval"`
	Small    int `form:"small" json:"small"`
	Big      int `form:"big" json:"big"`
}

var conf config

func (s *state) ToH() gin.H {
	return gin.H{
		"timer": s.Timer.Format(time.RFC3339),
		"small": s.Small,
		"big":   s.Big,
		"pause": s.Pause}
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

func loadConfig() error {
	if len(os.Args) == 1 {
		return fmt.Errorf("must provide config file path")
	}
	configFile := os.Args[1]

	contents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(contents, &conf)
}

func listenOnQueue() {
	ctx := context.TODO()

	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		panic("configuration error: " + err.Error())
	}

	client := sqs.NewFromConfig(cfg)

	urlResult, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: &conf.SqsQueue})
	if err != nil {
		panic("queue name error: " + err.Error())
	}
	queueURL := urlResult.QueueUrl

	receiveInput := &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
		QueueUrl:            queueURL,
		MaxNumberOfMessages: 1,
	}

	for {
		time.Sleep(time.Second * 1)

		res, err := client.ReceiveMessage(ctx, receiveInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "receive error: %s", err.Error())
			continue
		}

		if res.Messages == nil {
			continue
		}

		msg := res.Messages[0]
		var su stateUpdate
		err = json.Unmarshal([]byte(*msg.Body), &su)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unmarshal error: %s", err.Error())
			continue
		}
		s := state{
			Timer:    time.Now().Add(time.Minute * time.Duration(su.Interval)),
			Interval: su.Interval,
			Small:    su.Small,
			Big:      su.Big,
		}
		err = writeState(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "write error: %s", err.Error())
			continue
		}

		deleteInput := &sqs.DeleteMessageInput{
			QueueUrl:      queueURL,
			ReceiptHandle: msg.ReceiptHandle,
		}
		if _, err = client.DeleteMessage(ctx, deleteInput); err != nil {
			fmt.Fprintf(os.Stderr, "delete error: %s", err.Error())
			continue
		}
	}
}

func main() {
	err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if conf.SqsQueue != "" {
		go listenOnQueue()
	}

	router := gin.Default()
	router.StaticFile("/favicon.ico", "./public/favicon.ico")
	router.StaticFile("/", "./public/index.html")
	router.StaticFile("/fonts/Roboto-Thin.ttf", "./public/fonts/Roboto-Thin.ttf")
	router.GET("/state", getState)
	router.POST("/state", postState)
	router.POST("/pause", pause)
	router.POST("/resume", resume)
	router.Run(":8080")
}
