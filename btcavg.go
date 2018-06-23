package ticker

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

func NewBTCAVGFetcher(pubkey string, privkey string) fetchFn {
	return func() (exchangeRates, error) {
		output := exchangeRates{}
		errCh := make(chan error, 2)

		// Request both endpoints and save their responses
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			rates, err := fetchBTCAVGResource(btcavgFiatEndpoint, pubkey, privkey)
			if err != nil {
				errCh <- err
				return
			}

			formatBTCAVGFiatOutput(output, rates)
		}()

		go func() {
			defer wg.Done()
			rates, err := fetchBTCAVGResource(btcavgCryptoEndpoint, pubkey, privkey)
			if err != nil {
				errCh <- err
				return
			}

			err = formatBTCAVGCryptoOutput(output, rates)
			if err != nil {
				errCh <- err
				return
			}
		}()
		wg.Wait()
		close(errCh)

		for err := range errCh {
			return nil, err
		}

		return output, nil
	}
}

// fetchBTCAVGResource gets the response for a given BitcoinAverage endpoint
func fetchBTCAVGResource(url string, pubkey string, privkey string) (exchangeRates, error) {
	// Create signed request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-signature", createBTCAVGSignature(pubkey, privkey))

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

	// Return an error if no success, otherwise deserialize and return our body
	if resp.StatusCode != 200 {
		return nil, err
	}

	rates := make(exchangeRates)
	err = json.Unmarshal(body, &rates)
	if err != nil {
		return nil, err
	}

	return rates, nil
}

// formatBTCAVGFiatOutput formats BTC->fiat pairs
func formatBTCAVGFiatOutput(outgoing exchangeRates, incoming exchangeRates) {
	for k, v := range incoming {
		if strings.HasPrefix(k, "BTC") {
			outgoing[strings.TrimPrefix(k, "BTC")] = v
		}
	}
}

// formatBTCAVGCryptoOutput formats BTC->crypto pairs
func formatBTCAVGCryptoOutput(outgoing exchangeRates, incoming exchangeRates) error {
	for symbol, entry := range incoming {
		trimmedSymbol := strings.TrimSuffix(symbol, "BTC")
		if symbol == trimmedSymbol {
			continue
		}
		symbol := CanonicalizeSymbol(trimmedSymbol)

		if !IsCorrectIDForSymbol(symbol, entry.ID) {
			continue
		}

		if entry.Ask == "" || entry.Bid == "" || entry.Last == "" {
			continue
		}

		ask, err := invertAndFormatPrice(entry.Ask)
		if err != nil {
			return err
		}
		bid, err := invertAndFormatPrice(entry.Bid)
		if err != nil {
			return err
		}
		last, err := invertAndFormatPrice(entry.Last)
		if err != nil {
			return err
		}

		outgoing[symbol] = exchangeRate{Ask: ask, Bid: bid, Last: last}
	}
	return nil
}

func createBTCAVGSignature(pubkey string, privkey string) string {
	// Build payload
	payload := fmt.Sprintf("%d.%s", time.Now().Unix(), pubkey)

	// Generate the HMAC-sha256 signature
	mac := hmac.New(sha256.New, []byte(privkey))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	// Return the final payload
	return fmt.Sprintf("%s.%s", payload, signature)
}
