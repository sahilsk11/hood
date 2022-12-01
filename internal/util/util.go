package util

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func TimePtr(t time.Time) *time.Time {
	return &t
}

func StringPtr(s string) *string {
	return &s
}

type Secrets struct {
	AlphaVantageKey string `json:"alphaVantage"`
}

func LoadSecrets() (*Secrets, error) {
	f, err := os.ReadFile("secrets.json")
	if err != nil {
		return nil, fmt.Errorf("could not open secrets.json: %w", err)
	}
	secrets := Secrets{}
	err = json.Unmarshal(f, &secrets)
	if err != nil {
		return nil, err
	}

	return &secrets, nil
}
