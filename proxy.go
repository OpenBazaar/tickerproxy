package ticker

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gocraft/health"
)

// kvs is a helper type for logging
type kvs health.Kvs

type rateFetcher interface {
	fetch() (exchangeRates, error)
}

// exchangeRate represents the desired price data
type exchangeRate struct {
	Ask  json.Number `json:"ask"`
	Bid  json.Number `json:"bid"`
	Last json.Number `json:"last"`
	Type string      `json:"type"`
}

// exchangeRates represents a map of symbols to rate data for that symbol
type exchangeRates map[string]exchangeRate

// TestS3Region is the region to use in test mode
const TestS3Region = "test"

// Set package-level private vars
var (
	httpClient       = &http.Client{Timeout: 30 * time.Second}
	btcInvariantRate = exchangeRate{
		Ask:  "1",
		Bid:  "1",
		Last: "1",
		Type: exchangeRateTypeCrypto.String(),
	}
)

// Proxy gets data from the API endpoint and caches it
type Proxy struct {
	fetchers []rateFetcher

	// Output settings
	outfile  string
	s3Client *s3.S3
	s3Bucket string

	// Output state
	currentOutput []byte

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
	if s3Region != "" && s3Region != TestS3Region {
		creds := credentials.NewEnvCredentials()
		_, err := creds.Get()
		if err != nil {
			return nil, err
		}
		s3CFG := aws.NewConfig().WithRegion(s3Region).WithCredentials(creds) //.WithLogLevel(aws.LogDebug)
		s3Client = s3.New(session.New(), s3CFG)
	}

	return &Proxy{
		fetchers: []rateFetcher{
			newCMCFetcher(),
			newBTCAVGFetcher(pubkey, privkey),
		},

		outfile:  outfile,
		s3Client: s3Client,
		s3Bucket: s3Bucket,

		currentOutput: []byte("{}"),
		ticker:        time.NewTicker(time.Duration(speed) * time.Second),
		stream:        health.NewStream(),
		stopCh:        make(chan (struct{})),
		doneCh:        make(chan (struct{})),
	}, nil
}

// SetStream sets the health stream to write to
func (p *Proxy) SetStream(stream *health.Stream) {
	p.stream = stream
}

// Fetch gets data from all endpoints and caches them
func (p *Proxy) Fetch() error {
	job := p.stream.NewJob("proxy.fetch")

	// Fetch
	output, err := p.fetchAll()
	if err != nil {
		job.EventErr("fetch_all", err)
		job.Complete(health.Error)
		return err
	}
	p.currentOutput = output

	// Cache to disk
	if p.outfile != "" {
		err := ioutil.WriteFile(p.outfile, p.currentOutput, 0644)
		if err != nil {
			job.EventErr("write.outfile", err)
			job.Complete(health.Error)
			return err
		}
		job.EventKv("write.outfile", kvs{"file": p.outfile})
	}

	// Upload to AWS
	if p.s3Client != nil {
		err = sendToS3(p.s3Client, p.s3Bucket, p.currentOutput)
		if err != nil {
			job.EventErr("write.s3", err)
			job.Complete(health.Error)
			return err
		}
		job.Event("write.s3")
	}

	job.Complete(health.Success)

	return nil
}

// Start requests the latest data and begins polling every tick
func (p *Proxy) Start() {
	p.stream.Event("proxy.starting")

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
	return string(p.currentOutput)
}

func (p *Proxy) fetchAll() ([]byte, error) {
	allRates := []exchangeRates{}
	for _, f := range p.fetchers {
		rates, err := f.fetch()
		if err != nil {
			return nil, err
		}
		allRates = append(allRates, rates)
	}

	mergedRatesBytes, err := json.Marshal(mergeRates(allRates))
	if err != nil {
		return nil, err
	}

	return mergedRatesBytes, nil
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

func mergeRates(allRates []exchangeRates) exchangeRates {
	if len(allRates) == 0 {
		return nil
	}

	base := allRates[0]
	for _, rates := range allRates {
		for k, v := range rates {
			base[k] = v
		}
	}
	return base
}
