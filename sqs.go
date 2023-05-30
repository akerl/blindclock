package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type sqsClient struct {
	client   *sqs.Client
	queueURL string
}

func (s *sqsClient) getClient() (*sqs.Client, error) {
	if s.client == nil {
		ctx := context.TODO()

		cfg, err := awsConfig.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, err
		}

		s.client = sqs.NewFromConfig(cfg)
	}
	return s.client, nil
}

func (s *sqsClient) getQueueURL() (string, error) {
	if s.queueURL == "" {
		client, err := s.getClient()
		if err != nil {
			return "", err
		}
		urlResult, err := client.GetQueueUrl(
			context.TODO(),
			&sqs.GetQueueUrlInput{QueueName: &conf.SqsQueue},
		)
		if err != nil {
			return "", err
		}
		s.queueURL = *urlResult.QueueUrl
	}
	return s.queueURL, nil
}

func (s *sqsClient) Receive() (*types.Message, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}

	queueURL, err := s.getQueueURL()
	if err != nil {
		return nil, err
	}

	res, err := client.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{QueueUrl: &queueURL})
	if err != nil {
		return nil, err
	}

	if res.Messages == nil {
		return nil, nil
	}

	return &res.Messages[0], nil
}

func (s *sqsClient) Delete(rh string) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}

	queueURL, err := s.getQueueURL()
	if err != nil {
		return err
	}

	deleteInput := &sqs.DeleteMessageInput{
		QueueUrl:      &queueURL,
		ReceiptHandle: &rh,
	}
	_, err = client.DeleteMessage(context.TODO(), deleteInput)
	return err
}

func sqsUpdate(m types.Message) error { //revive:disable-line:cyclomatic
	var su stateUpdate
	if err := json.Unmarshal([]byte(*m.Body), &su); err != nil {
		return err
	}

	s, err := readState()
	if err != nil {
		return err
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

func listenOnQueue() {
	sc := sqsClient{}

	for {
		time.Sleep(time.Second * 1)

		msg, err := sc.Receive()
		if err != nil {
			fmt.Fprintf(os.Stderr, "receive error: %s", err.Error())
			continue
		}
		if msg == nil {
			continue
		}

		if err := sqsUpdate(*msg); err != nil {
			fmt.Fprintf(os.Stderr, "update error: %s", err.Error())
			continue
		}

		if err := sc.Delete(*msg.ReceiptHandle); err != nil {
			fmt.Fprintf(os.Stderr, "delete error: %s", err.Error())
		}
	}
}
