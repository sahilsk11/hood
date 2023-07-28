package util

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

var EnableDebug = false

func TimePtr(t time.Time) *time.Time {
	return &t
}

func StringPtr(s string) *string {
	return &s
}

type Secrets struct {
	AlphaVantageKey string `json:"alphaVantage"`
	RdsPassword     string `json:"db"`
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

func Pprint(i interface{}) {
	if EnableDebug {
		bytes, err := json.MarshalIndent(i, "", "    ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(bytes))
	}
}

func SortedMapKeys[T any](in map[string]T) []string {
	out := make([]string, 0, len(in))
	for s := range in {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
