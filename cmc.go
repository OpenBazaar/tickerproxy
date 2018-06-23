package ticker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const (
	cmcQueryEndpointTempalte = "https://api.coinmarketcap.com/v2/ticker?convert=BTC&start=%d&limit=%d"
	cmcQueryFirstID          = 1
	defaultCMCQueryLimit     = 100
)

type cmcCoinData struct {
	ID     int64  `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
	Quotes struct {
		BTC struct {
			Price     json.Number `json:"price"`
			MarketCap float64     `json:"market_cap"`
		} `json:"BTC"`
	} `json:"quotes"`
}

type cmcResponse struct {
	Data map[json.Number]cmcCoinData `json:"data"`

	Metadata struct {
		Count int `json:"num_cryptocurrencies"`
	} `json:"metadata"`
}

func FetchCMC() (exchangeRates, error) {
	output := exchangeRates{}

	resp, err := fetchCMCResource(cmcQueryFirstID, output)
	if err != nil {
		return nil, err
	}

	for i := defaultCMCQueryLimit + cmcQueryFirstID; i < resp.Metadata.Count; i += defaultCMCQueryLimit {
		_, err = fetchCMCResource(i, output)
		if err != nil {
			return nil, err
		}
	}

	return output, nil
}

func fetchCMCResource(start int, output exchangeRates) (*cmcResponse, error) {
	resp, err := httpClient.Get(buildCMCQueryEndpoint(start, defaultCMCQueryLimit))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	payload := &cmcResponse{}
	err = json.Unmarshal(body, payload)
	if err != nil {
		return nil, err
	}

	for _, entry := range payload.Data {
		entry.Symbol = CanonicalizeSymbol(entry.Symbol)

		if !IsCorrectIDForSymbol(entry.Symbol, entry.ID) {
			continue
		}

		if entry.Quotes.BTC.Price == "" {
			continue
		}

		price, err := invertAndFormatPrice(entry.Quotes.BTC.Price)
		if err != nil {
			return nil, err
		}

		output[entry.Symbol] = exchangeRate{Ask: price, Bid: price, Last: price}
	}

	return payload, nil
}

func buildCMCQueryEndpoint(start int, limit int) string {
	return fmt.Sprintf(cmcQueryEndpointTempalte, start, limit)
}
