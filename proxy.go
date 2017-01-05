package tickerproxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"strconv"

	"github.com/gocraft/health"
)

// tickerEndpoint is the default API endpoint to proxy
const tickerEndpoint = "https://api.bitcoinaverage.com/ticker/global/all"

// httpClient is a an http client with a read timeout set
var httpClient = &http.Client{Timeout: 10 * time.Second}

// kvs is a helper type for logging
type kvs health.Kvs

// Proxy gets data from the API endpoint and caches it
type Proxy struct {
	// Settings
	url        string
	publicKey  string
	privateKey string
	speed      int

	// Proxied data
	lastResponseBody    []byte
	lastResponseHeaders http.Header

	// Mechanics
	ticker *time.Ticker
	stream *health.Stream

	stopCh chan (struct{})
	doneCh chan (struct{})
}

// New creates a new `Proxy` with the given credentials and default values
func New(speed int, pubkey string, privkey string) *Proxy {
	return &Proxy{
		// Settings
		url:        tickerEndpoint + "?public_key=" + pubkey,
		publicKey:  pubkey,
		privateKey: privkey,
		speed:      speed,

		// Initial data
		lastResponseBody: []byte("{}"),

		// Mechanics
		ticker: time.NewTicker(time.Duration(speed) * time.Second),
		stream: health.NewStream(),

		stopCh: make(chan (struct{})),
		doneCh: make(chan (struct{})),
	}
}

// SetStream sets the health stream to write to
func (p *Proxy) SetStream(stream *health.Stream) {
	p.stream = stream
}

// Fetch gets the latest data from the API server
func (p *Proxy) Fetch() error {
	job := p.stream.NewJob("proxy.fetch")

	// Create the http request
	req, err := http.NewRequest("GET", p.url, nil)
	if err != nil {
		job.EventErr("request.new", err)
		job.Complete(health.Error)
		return err
	}
	req.Header.Set("X-Signature", p.currentSignature())

	// Send the request
	resp, err := httpClient.Get(p.url)
	if err != nil {
		job.EventErr("request.make", err)
		job.Complete(health.Error)
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		job.EventErr("request.read", err)
		job.Complete(health.Error)
		return err
	}

	// Update cache on success, or notify upon failure
	if resp.StatusCode == 200 {
		p.lastResponseBody = body
		p.lastResponseHeaders = resp.Header
	} else {
		job.EventErr("request.status", errors.New(string(body)))
	}

	job.Complete(health.Success)

	return nil
}

// Start requests the latest data and begins polling every tick
func (p *Proxy) Start() {
	p.stream.EventKv("proxy.starting", kvs{"public_key": p.publicKey, "speed": strconv.Itoa(p.speed)})

	p.Fetch()
	for {
		select {
		case <-p.stopCh:
			close(p.doneCh)
			return
		case <-p.ticker.C:
			p.Fetch()
		}
	}
}

// Stop makes the `Proxy` stop fetching new data
func (p *Proxy) Stop() {
	p.stream.Event("proxy.stopping")
	close(p.stopCh)
	<-p.doneCh
}

// String returns the latest response from the API as a string
func (p *Proxy) String() string {
	return string(p.lastResponseBody)
}

// ServeHTTP is an `http.Handler` that just returns the lastest response from
// the upstream server
func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	job := p.stream.NewJob("proxy.serve")

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
	payload := fmt.Sprintf("%d.%s", time.Now().Unix(), p.publicKey)

	// Generate the HMAC-sha256 signature
	// As per the docs, do not decode the key base64, but do encode the output
	mac := hmac.New(sha256.New, []byte(p.privateKey))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	// Return the final payload
	return fmt.Sprintf("%s.%s", payload, signature)
}
