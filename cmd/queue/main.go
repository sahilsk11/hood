package main

import (
	"context"
	"database/sql"
	"log"

	"hood/internal/queue"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	_ "github.com/lib/pq"
)

func main() {
	connStr := "postgresql://postgres:postgres@localhost:5438/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.WithValue(
		context.Background(),
		"tx",
		tx,
	)

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

	err = queue.GetAndProcess(ctx, sqsService)
	if err != nil {
		log.Fatal(err)
	}
}
