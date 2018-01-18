package tickerproxy

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gocraft/health"
)

// kvs is a helper type for logging
type kvs health.Kvs

// TestS3Region is the region to use in test mode
const TestS3Region = "test"

// Set package-level private vars
var (
	altcoins         = []string{"BCH", "ZEC", "LTC", "XMR", "ETH"}
	fiatEndpoint     = "https://apiv2.bitcoinaverage.com/indices/global/ticker/all?crypto=BTC"
	cryptoEndpoint   = "https://apiv2.bitcoinaverage.com/indices/crypto/ticker/all?crypto=BTC," + strings.Join(altcoins, ",")
	httpClient       = &http.Client{Timeout: 10 * time.Second}
	btcInvariantRate = exchangeRate{
		Ask:  "1",
		Bid:  "1",
		Last: "1",
	}
)

type exchangeRate struct {
	Ask  json.Number `json:"ask"`
	Bid  json.Number `json:"bid"`
	Last json.Number `json:"last"`
}

type exchangeRates map[string]exchangeRate

// Proxy gets data from the API endpoint and caches it
type Proxy struct {
	// Settings
	fiatURL   string
	cryptoURL string

	outfile    string
	publicKey  string
	privateKey string
	speed      int

	// S3
	s3Client *s3.S3
	s3Bucket string

	// Proxied data
	lastResponseBody []byte

	// Mechanics
	ticker *time.Ticker
	stream *health.Stream

	stopCh chan (struct{})
	doneCh chan (struct{})
}

// New creates a new `Proxy` with the given credentials and default values
func New(speed int, pubkey string, privkey string, outfile string, s3Region string, s3Bucket string) (*Proxy, error) {
	var s3Client *s3.S3

	// Configure S3 client unless we're in test mode
	if s3Region != TestS3Region {
		creds := credentials.NewEnvCredentials()
		_, err := creds.Get()
		if err != nil {
			return nil, err
		}
		s3CFG := aws.NewConfig().WithRegion(s3Region).WithCredentials(creds) //.WithLogLevel(aws.LogDebug)
		s3Client = s3.New(session.New(), s3CFG)
	}

	return &Proxy{
		// Settings
		fiatURL:    fiatEndpoint,
		cryptoURL:  cryptoEndpoint,
		outfile:    outfile,
		publicKey:  pubkey,
		privateKey: privkey,
		speed:      speed,

		s3Client: s3Client,
		s3Bucket: s3Bucket,

		// Initial data
		lastResponseBody: []byte("{}"),

		// Mechanics
		ticker: time.NewTicker(time.Duration(speed) * time.Second),
		stream: health.NewStream(),

		stopCh: make(chan (struct{})),
		doneCh: make(chan (struct{})),
	}, nil
}

// SetStream sets the health stream to write to
func (p *Proxy) SetStream(stream *health.Stream) {
	p.stream = stream
}

// Fetch gets the latest data from the API server
func (p *Proxy) Fetch() error {
	job := p.stream.NewJob("proxy.fetch")

	// Request both endpoints and save their responses
	httpReqs := map[string][]byte{
		p.fiatURL:   nil,
		p.cryptoURL: nil,
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	for url := range httpReqs {
		go func(url string) {
			defer func() {
				wg.Done()
			}()

			kvs := kvs{"url": url}

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				job.EventErrKv("request.new", err, kvs)
				job.Complete(health.Error)
				return
			}
			req.Header.Add("X-signature", createSignature(p.publicKey, p.privateKey))

			// Send the requests
			resp, err := httpClient.Do(req)
			if err != nil {
				job.EventErrKv("request.make", err, kvs)
				job.Complete(health.Error)
				return
			}
			defer resp.Body.Close()

			// Read the response body
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				job.EventErrKv("request.read", err, kvs)
				job.Complete(health.Error)
				return
			}

			if resp.StatusCode != 200 {
				job.EventErrKv("request.status", errors.New(string(body)), kvs)
				job.Complete(health.Error)
				return
			}

			// Save to map
			httpReqs[url] = body
		}(url)
	}
	wg.Wait()

	// Save headers and formatted response
	var err error
	p.lastResponseBody, err = formatResponse(httpReqs[p.fiatURL], httpReqs[p.cryptoURL])
	if err != nil {
		job.EventErr("request.format_response", err)
		job.Complete(health.Error)
		return err
	}

	// Cache to disk
	if p.outfile != "" {
		err := ioutil.WriteFile(p.outfile, p.lastResponseBody, 0644)
		if err != nil {
			job.EventErr("request.write_to_outfile", err)
			job.Complete(health.Error)
			return err
		}
		job.EventKv("request.write_to_outfile", kvs{"file": p.outfile})
	}

	// Upload to AWS
	if p.s3Client != nil {
		err = sendToS3(p.s3Client, p.s3Bucket, p.lastResponseBody)
		if err != nil {
			job.EventErr("aws.write", err)
			job.Complete(health.Error)
			return err
		}
		job.Event("aws.write")
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

	// Write body
	_, err := w.Write(p.lastResponseBody)
	if err != nil {
		job.EventErr("write", err)
		job.Complete(health.Error)
	}

	job.Complete(health.Success)
}

// createSignature generates a bitcoinaverage.com API signature
func createSignature(publicKey string, privateKey string) string {
	// Build payload
	payload := fmt.Sprintf("%d.%s", time.Now().Unix(), publicKey)

	// Generate the HMAC-sha256 signature
	// As per the docs, do not decode the key base64, but do encode the output
	mac := hmac.New(sha256.New, []byte(privateKey))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	// Return the final payload
	return fmt.Sprintf("%s.%s", payload, signature)
}

func formatResponse(fiatBody []byte, cryptoBody []byte) ([]byte, error) {
	// Read API responses into memory
	incomingFiat := make(exchangeRates)
	incomingCrypto := make(exchangeRates)

	if fiatBody != nil {
		err := json.Unmarshal(fiatBody, &incomingFiat)
		if err != nil {
			return nil, err
		}
	}

	if cryptoBody != nil {
		err := json.Unmarshal(cryptoBody, &incomingCrypto)
		if err != nil {
			return nil, err
		}
	}

	// Prepare a new outgoing response
	outgoing := make(exchangeRates, len(incomingFiat)+len(incomingCrypto)+1)
	outgoing["BTC"] = btcInvariantRate

	// Add data from fiat response
	for k, v := range incomingFiat {
		if strings.HasPrefix(k, "BTC") {
			outgoing[strings.TrimPrefix(k, "BTC")] = v
		}
	}

	// Add data from crypto response
	// Crypto pairs need to be reversed into BTC-<x> pairs
	for k, v := range incomingCrypto {
		for _, altcoinSymbol := range altcoins {
			if k == altcoinSymbol+"BTC" {
				ask, err := v.Ask.Float64()
				if err != nil {
					return nil, err
				}
				bid, err := v.Bid.Float64()
				if err != nil {
					return nil, err
				}
				last, err := v.Last.Float64()
				if err != nil {
					return nil, err
				}

				ask = 1.0 / ask
				bid = 1.0 / bid
				last = 1.0 / last

				outgoing[altcoinSymbol] = exchangeRate{
					Ask:  json.Number(strconv.FormatFloat(ask, 'G', -1, 32)),
					Bid:  json.Number(strconv.FormatFloat(bid, 'G', -1, 32)),
					Last: json.Number(strconv.FormatFloat(last, 'G', -1, 32)),
				}
			}
		}
	}

	// Serialize the formatted response
	outgoingBytes, err := json.Marshal(outgoing)
	if err != nil {
		return nil, err
	}

	return outgoingBytes, nil
}

func sendToS3(s3Client *s3.S3, bucket string, data []byte) error {
	_, err := s3Client.PutObject(&s3.PutObjectInput{
		Key:           aws.String("api"),
		Bucket:        aws.String(bucket),
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
		ContentType:   aws.String("application/json"),
	})
	if err != nil {
		return err
	}

	return nil
}
