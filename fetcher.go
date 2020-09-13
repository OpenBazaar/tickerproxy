package ticker

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gocraft/health"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

type fetchFn func() (exchangeRates, error)

// Fetch gets data from all sources, formats it, and sends it to the Writers.
func Fetch(stream *health.Stream, conf Config, writers ...Writer) error {
	job := stream.NewJob("fetch")

	// Fetch data from each provider
	allRates := []exchangeRates{{"BTC": {Ask: "1", Bid: "1", Last: "1", Type: exchangeRateTypeCrypto.String()}}}
	for _, f := range []fetchFn{
		NewBTCAVGFetcher(conf.BTCAVGPubkey, conf.BTCAVGPrivkey),
		NewCMCFetcher(conf.CMCEnv, conf.CMCAPIKey),
	} {
		rates, err := f()
		if err != nil {
			job.EventErr("fetch_data", err)
			job.Complete(health.Error)
			return err
		}
		allRates = append(allRates, rates)
	}

	fullRates := mergeRates(allRates)

	// Ensure the final payload passes correctness checks
	err := validateRates(fullRates)
	if err != nil {
		job.EventErr("validate_rates", err)
		job.Complete(health.Error)
		return err
	}

	// Serialize responses
	responseBytes, err := json.Marshal(fullRates)
	if err != nil {
		job.EventErr("marshal", err)
		job.Complete(health.Error)
		return err
	}

	// Write
	for _, writer := range writers {
		err := writer(job, responseBytes)
		if err != nil {
			job.EventErr("write", err)
			job.Complete(health.Error)
			return err
		}
	}

	job.Complete(health.Success)
	return nil
}

func validateRates(rates exchangeRates) error {
	for _, symbol := range RequiredSymbols {
		if _, ok := rates[symbol]; !ok {
			return errFetchMissingRequiredSymbol(symbol)
		}
	}

	return nil
}

type errFetchMissingRequiredSymbol string

func (e errFetchMissingRequiredSymbol) Error() string {
	return "Missing required symbol: " + string(e)
}
