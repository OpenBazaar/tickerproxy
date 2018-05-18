package ticker

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	btcavgFiatEndpoint   = "https://apiv2.bitcoinaverage.com/indices/global/ticker/all?crypto=BTC"
	btcavgCryptoEndpoint = "https://apiv2.bitcoinaverage.com/indices/crypto/ticker/all"
)

type btcavgFetcher struct {
	pubkey  string
	privkey string
}

func newBTCAVGFetcher(pubkey string, privkey string) *btcavgFetcher {
	return &btcavgFetcher{
		pubkey:  pubkey,
		privkey: privkey,
	}
}

func (f *btcavgFetcher) fetch() (exchangeRates, error) {
	output := exchangeRates{"BTC": btcInvariantRate}

	// Request both endpoints and save their responses
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		body, err := f.fetchResponse(btcavgFiatEndpoint)
		if err != nil {
			return
		}

		incoming := make(exchangeRates)
		err = json.Unmarshal(body, &incoming)
		if err != nil {
			return
		}

		formatFiatOutput(output, incoming)
	}()

	go func() {
		defer wg.Done()
		body, err := f.fetchResponse(btcavgCryptoEndpoint)
		if err != nil {
			return
		}
		incoming := make(exchangeRates)
		err = json.Unmarshal(body, &incoming)
		if err != nil {
			return
		}

		err = formatCryptoOutput(output, incoming)
		if err != nil {
			return
		}
	}()
	wg.Wait()

	return output, nil
}

// fetchResponse gets the response for a given BitcoinAverage endpoint
func (f *btcavgFetcher) fetchResponse(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-signature", f.signature())
	// Send the requests
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Return an error if no success, otherwise return our body
	if resp.StatusCode != 200 {
		return nil, err
	}
	return body, nil
}

// formatFiatOutput formats BTC->fiat pairs
func formatFiatOutput(outgoing exchangeRates, incoming exchangeRates) {
	for k, v := range incoming {
		v.Type = exchangeRateTypeFiat.String()
		if strings.HasPrefix(k, "BTC") {
			outgoing[strings.TrimPrefix(k, "BTC")] = v
		}
	}
}

// formatCryptoOutput formats BTC->crypto pairs
func formatCryptoOutput(outgoing exchangeRates, incoming exchangeRates) error {
	var altcoinSymbol string
	for k, v := range incoming {
		if !strings.HasSuffix(k, "BTC") {
			continue
		}

		altcoinSymbol = strings.TrimSuffix(k, "BTC")

		ask, err := v.Ask.Float64()
		if err != nil {
			return err
		}
		bid, err := v.Bid.Float64()
		if err != nil {
			return err
		}
		last, err := v.Last.Float64()
		if err != nil {
			return err
		}

		ask = 1.0 / ask
		bid = 1.0 / bid
		last = 1.0 / last

		outgoing[altcoinSymbol] = exchangeRate{
			Type: exchangeRateTypeCrypto.String(),
			Ask:  json.Number(strconv.FormatFloat(ask, 'G', -1, 32)),
			Bid:  json.Number(strconv.FormatFloat(bid, 'G', -1, 32)),
			Last: json.Number(strconv.FormatFloat(last, 'G', -1, 32)),
		}
	}
	return nil
}

func (f *btcavgFetcher) signature() string {
	// Build payload
	payload := fmt.Sprintf("%d.%s", time.Now().Unix(), f.pubkey)

	// Generate the HMAC-sha256 signature
	mac := hmac.New(sha256.New, []byte(f.privkey))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	// Return the final payload
	return fmt.Sprintf("%s.%s", payload, signature)
}
