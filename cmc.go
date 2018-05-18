package ticker

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	cmcBaseEndpoint = "https://api.coinmarketcap.com/v2/ticker?convert=BTC&limit=100&start="
)

type cmcFetcher struct {
	pubkey  string
	privkey string
}

type cmcResponse struct {
	Data map[string]struct {
		Symbol string `json:"symbol"`
		Quotes struct {
			BTC struct {
				Price float64 `json:"price"`
			} `json:"BTC"`
		} `json:"quotes"`
	} `json:"data"`

	Metadata struct {
		Count int `json:"num_cryptocurrencies"`
	} `json:"metadata"`
}

func newCMCFetcher() *cmcFetcher {
	return &cmcFetcher{}
}

func (f *cmcFetcher) fetch() (exchangeRates, error) {
	output := exchangeRates{}

	resp, err := f.fetchPage(1)
	if err != nil {
		return nil, err
	}
	addPage(output, resp)

	for i := 101; i < resp.Metadata.Count; i += 100 {
		resp, err := f.fetchPage(i)
		if err != nil {
			return nil, err
		}

		if len(resp.Data) == 0 {
			break
		}

		addPage(output, resp)
	}

	return output, nil
}

func (f *cmcFetcher) fetchPage(start int) (*cmcResponse, error) {
	resp, err := http.Get(cmcBaseEndpoint + strconv.Itoa(start))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, err
	}

	data := &cmcResponse{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func addPage(rates exchangeRates, pageResp *cmcResponse) {
	for _, entry := range pageResp.Data {
		formattedPrice := json.Number(strconv.FormatFloat(entry.Quotes.BTC.Price, 'G', -1, 64))
		rates[entry.Symbol] = exchangeRate{
			Type: exchangeRateTypeCrypto.String(),
			Ask:  formattedPrice,
			Bid:  formattedPrice,
			Last: formattedPrice,
		}
	}
}
