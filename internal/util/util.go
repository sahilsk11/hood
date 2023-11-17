package util

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/plaid/plaid-go/plaid"
)

var EnableDebug = true

func TimePtr(t time.Time) *time.Time {
	return &t
}

func StringPtr(s string) *string {
	return &s
}

type Secrets struct {
	AlphaVantageKey   string `json:"alphaVantage"`
	RdsPassword       string `json:"db"`
	DataJockeryApiKey string `json:"dataJockey"`
	Plaid             struct {
		ClientID string `json:"clientID"`
		Secret   string `json:"secret"`
	} `json:"plaid"`
}

type Environment string

const Production Environment = "prod"
const Development Environment = "dev"

func (e Environment) ToPlaidEnv() plaid.Environment {
	return map[Environment]plaid.Environment{
		Production:  plaid.Production,
		Development: plaid.Sandbox,
	}[e]
}

func LoadSecrets(env Environment) (*Secrets, error) {
	f, err := os.ReadFile("secrets.json")
	if err != nil {
		return nil, fmt.Errorf("could not open secrets.json: %w", err)
	}
	secrets := map[Environment]*Secrets{}
	err = json.Unmarshal(f, &secrets)
	if err != nil {
		return nil, err
	}

	return secrets[env], nil
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
