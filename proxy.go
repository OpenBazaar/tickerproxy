package tickerproxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gocraft/health"
)

// DefaultTickerEndpoint is the default API endpoint to proxy
const DefaultTickerEndpoint = "https://api.bitcoinaverage.com/ticker/global/all"

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type kvs map[string]string

// Proxy gets data from the API endpoint and caches it
type Proxy struct {
	URL        string
	PublicKey  string
	PrivateKey string

	lastResponseBody    []byte
	lastResponseHeaders http.Header

	ticker *time.Ticker
	stream *health.Stream
}

// New creates a new `Proxy` with the given credentials
func New(pubkey string, privkey string) *Proxy {
	stream := health.NewStream()
	stream.AddSink(&health.WriterSink{os.Stdout})

	return &Proxy{
		URL:        DefaultTickerEndpoint + "?public_key=" + pubkey,
		PublicKey:  pubkey,
		PrivateKey: privkey,

		ticker: time.NewTicker(10 * time.Second),
		stream: stream,
	}
}

// Fetch gets the latest data from the API server
func (p *Proxy) Fetch() error {
	// Create a health job for the fetch,
	job := p.stream.NewJob("fetch")
	job.KeyValue("url", p.URL)

	// Create the http request
	req, err := http.NewRequest("GET", p.URL, nil)
	if err != nil {
		job.EventErr("new_request", err)
		job.Complete(health.Error)
		return err
	}
	req.Header.Set("X-signature", p.currentSignature())

	// Send the request
	resp, err := httpClient.Get(p.URL)
	if err != nil {
		job.EventErr("get", err)
		job.Complete(health.Error)
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		job.EventErr("read_all", err)
		job.Complete(health.Error)
		return err
	}

	// Update cache
	p.lastResponseBody = body
	p.lastResponseHeaders = resp.Header

	job.Complete(health.Success)
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
	job := p.stream.NewJob("serve")

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
		job.EventErr("write", err)
		job.Complete(health.Error)
	}

	job.Complete(health.Success)
}

// currentSignature generates a bitcoinaverage.com API signature
func (p *Proxy) currentSignature() string {
	// Build payload
	payload := fmt.Sprintf("%d.%s", time.Now().Unix(), p.PublicKey)

	// Generate the HMAC-sha256 signature
	mac := hmac.New(sha256.New, []byte(p.PrivateKey))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	// Return the final payload
	return fmt.Sprintf("%s.%s", payload, signature)
}
