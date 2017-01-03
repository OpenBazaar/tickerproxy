package tickerproxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// tickerEndpoint is the default API endpoint to proxy
const tickerEndpoint = "https://api.bitcoinaverage.com/ticker/global/all"

// httpClient is a an http client with a read timeout set
var httpClient = &http.Client{Timeout: 10 * time.Second}

// Proxy gets data from the API endpoint and caches it
type Proxy struct {
	url                 string
	publicKey           string
	privateKey          string
	lastResponseBody    []byte
	lastResponseHeaders http.Header
	ticker              *time.Ticker
}

// New creates a new `Proxy` with the given credentials
func New(speed time.Duration, pubkey string, privkey string) *Proxy {
	return &Proxy{
		url:        tickerEndpoint + "?public_key=" + pubkey,
		publicKey:  pubkey,
		privateKey: privkey,
		ticker:     time.NewTicker(speed),
	}
}

// Fetch gets the latest data from the API server
func (p *Proxy) Fetch() error {
	// Create the http request
	req, err := http.NewRequest("GET", p.url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Signature", p.currentSignature())

	// Send the request
	resp, err := httpClient.Get(p.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Update cache
	p.lastResponseBody = body
	p.lastResponseHeaders = resp.Header

	return nil
}

// Start requests the latest data and begins polling every tick
func (p *Proxy) Start() {
	p.Fetch()
	for range p.ticker.C {
		p.Fetch()
	}
}

// Stop makes the `Proxy` stop fetching new data
func (p *Proxy) Stop() {
	p.ticker.Stop()
}

// String returns the latest response from the API as a string
func (p *Proxy) String() string {
	return string(p.lastResponseBody)
}

// ServeHTTP is an `http.Handler` that just returns the lastest response from
// the upstream server
func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Add headers
	for key, vals := range p.lastResponseHeaders {
		for _, val := range vals {
			w.Header().Set(key, val)
		}
	}
	w.Header().Set("X-OpenBazaar", "Trade free!")

	// Write body
	_, err := w.Write(p.lastResponseBody)
	if err != nil {
		fmt.Println(err)
	}
}

// currentSignature generates a bitcoinaverage.com API signature
func (p *Proxy) currentSignature() string {
	// Build payload
	payload := fmt.Sprintf("%d.%s", time.Now().Unix(), p.publicKey)

	// Generate the HMAC-sha256 signature
	// As per the docs, do not decode the key base64, but do encode the output
	mac := hmac.New(sha256.New, []byte(p.privateKey))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	// Return the final payload
	return fmt.Sprintf("%s.%s", payload, signature)
}
