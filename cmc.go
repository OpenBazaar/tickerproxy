package ticker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	cmcQueryEndpointTemplate = "https://%s-api.coinmarketcap.com/v1/cryptocurrency/listings/latest"
	cmcQueryFirstID          = 1
)

var cmcQueryLimit = 5000

var bannedCryptoSymbols = map[string]struct{}{
	"CRC": struct{}{},
	"HCA": struct{}{},
}

type cmcResponse struct {
	Data []struct {
		ID     int64  `json:"id"`
		Symbol string `json:"symbol"`
		Name   string `json:"name"`
		Quote  struct {
			BTC struct {
				Price json.Number `json:"price"`
			} `json:"BTC"`
		} `json:"quote"`
	} `json:"data"`
}

func NewCMCFetcher(env string, apiKey string) fetchFn {
	return func() (exchangeRates, error) {
		var (
			err    error = nil
			resp         = &cmcResponse{}
			output       = exchangeRates{}
		)

		// Start at the first ID and keep grabbing pages until we get less than we
		// requested or there is an error
		for i := 0; i < 100; i++ {
			resp, err = fetchCMCResource(env, apiKey, cmcQueryFirstID+(i*cmcQueryLimit), cmcQueryLimit, output)
			if err != nil {
				return nil, err
			}

			// We aren't getting any more data; stop
			if len(resp.Data) < cmcQueryLimit {
				break
			}
		}

		return output, nil
	}
}

func fetchCMCResource(host string, apiKey string, start int, limit int, output exchangeRates) (*cmcResponse, error) {
	req, err := http.NewRequest("GET", buildCMCEndpoint(host), nil)
	if err != nil {
		return nil, err
	}

	startStr := fmt.Sprintf("%v", start)
	limitStr := fmt.Sprintf("%v", limit)

	q := url.Values{}
	q.Add("start", startStr)
	q.Add("limit", limitStr)
	q.Add("convert", "BTC")

	req.Header.Add("X-CMC_PRO_API_KEY", apiKey)
	req.Header.Set("Accepts", "application/json")
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
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

		// Remove symbols that we don't want included in the API
		if _, ok := bannedCryptoSymbols[entry.Symbol]; ok {
			continue
		}

		if !IsCorrectIDForSymbol(entry.Symbol, entry.ID) {
			continue
		}

		price, err := invertAndFormatPrice(entry.Quote.BTC.Price)
		if err != nil {
			return nil, err
		}

		output[entry.Symbol] = exchangeRate{
			Ask:  price,
			Bid:  price,
			Last: price,
			Type: exchangeRateTypeCrypto.String(),
		}
	}

	return payload, nil
}

func buildCMCEndpoint(env string) string {
	return fmt.Sprintf(cmcQueryEndpointTemplate, env)
}
