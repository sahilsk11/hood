package queue

import (
	"context"
	"encoding/json"
	"fmt"
	data_ingestion "hood/internal/data-ingestion"
	"hood/internal/db/models/postgres/public/model"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/shopspring/decimal"
)

type tradeMessage struct {
	Action    string `json:"action"`
	Symbol    string `json:"symbol"`
	Quantity  string `json:"quantity"`
	Date      string `json:"date"`
	CostBasis string `json:"cost_basis"`
}

func GetAndProcess(ctx context.Context, sqsService *sqs.SQS) error {
	tradeMessage, err := getNext(sqsService)
	if err != nil {
		return nil
	}
	trade := model.Trade{
		Symbol: tradeMessage.Symbol,
	}
	if tradeMessage.Action == "SELL" {
		trade.Action = model.TradeActionType_Sell
	} else if tradeMessage.Action == "BUY" {
		trade.Action = model.TradeActionType_Buy
	} else {
		return fmt.Errorf("invalid trade action %s", tradeMessage.Action)
	}
	quantity, err := decimal.NewFromString(tradeMessage.Quantity)
	if err != nil {
		return fmt.Errorf("could not parse trade quantity: %w", err)
	}
	trade.Quantity = quantity
	costBasis, err := decimal.NewFromString(tradeMessage.CostBasis)
	if err != nil {
		return fmt.Errorf("could not parse cost basis: %w", err)
	}
	trade.CostBasis = costBasis

	et, err := time.LoadLocation("America/New_York")
	if err != nil {
		return err
	}
	date, err := time.ParseInLocation("2006-01-02 15:04:05", tradeMessage.Date, et)
	if err != nil {
		return err
	}
	trade.Date = date

	if trade.Action == model.TradeActionType_Buy {
		_, _, err = data_ingestion.AddBuyOrder(ctx, trade)
	} else {
		_, _, err = data_ingestion.AddSellOrder(ctx, trade)
	}

	return err
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
