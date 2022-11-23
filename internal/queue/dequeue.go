package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String("us-east-1"),
			CredentialsChainVerboseErrors: aws.Bool(true)},
		Profile: "hood-app-readonly",
	})
	if err != nil {
		log.Fatal(err)
	}
	_, err = sess.Config.Credentials.Get()
	if err != nil {
		log.Fatal(err)
	}
	sqsService := sqs.New(sess)

	out, err := getNext(sqsService)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)
}

type tradeMessage struct {
	Action   string `json:"action"`
	Symbol   string `json:"symbol"`
	Quantity string `json:"quantity"`
	Datr     string `json:"date"`
}

func getNext(sqsService *sqs.SQS) (*tradeMessage, error) {

	url := "https://sqs.us-east-1.amazonaws.com/326651360928/hood-email-queue"
	out, err := sqsService.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl: &url,
	})
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, fmt.Errorf("no result in queue")
	}
	if len(out.Messages) == 0 {
		return nil, fmt.Errorf("no new messages in result")
	}

	message := *out.Messages[0].Body
	var trade tradeMessage
	err = json.Unmarshal([]byte(message), &trade)
	if err != nil {
		return nil, err
	}

	return &trade, nil
}
