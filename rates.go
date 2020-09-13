package ticker

import (
	"encoding/json"
	"strconv"
)

// exchangeRate represents the desired price data
type exchangeRate struct {
	ID   int64       `json:"id,omitempty"`
	Ask  json.Number `json:"ask"`
	Bid  json.Number `json:"bid"`
	Last json.Number `json:"last"`
	Type string      `json:"type"`
}

// exchangeRates represents a map of symbols to rate data for that symbol
type exchangeRates map[string]exchangeRate

type rateFetcher interface {
	fetch() (exchangeRates, error)
}

func mergeRates(allRates []exchangeRates) exchangeRates {
	if len(allRates) == 0 {
		return nil
	}

	base := allRates[0]
	if len(allRates) == 1 {
		return base
	}

	for _, rates := range allRates[1:] {
		for k, v := range rates {
			base[k] = v
		}
	}
	return base
}

func invertAndFormatPrice(price json.Number) (json.Number, error) {
	if price == "" {
		return "", nil
	}
	priceAsFloat, err := price.Float64()
	if err != nil {
		return "", err
	}

	if priceAsFloat == 0 {
		return json.Number("0"), nil
	}
	return json.Number(strconv.FormatFloat(1.0/priceAsFloat, 'f', -1, 32)), nil
}
