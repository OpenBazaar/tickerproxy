package tickerproxy

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gocraft/health"
	"github.com/jarcoal/httpmock"
)

// const testResponseBody = `{"BTC": {"low": 0, "high": 1200.00}}`
const testResponseBody = `{
  "BTCUSD": {
    "ask": "22324.79",
    "bid": "22299.15",
		"last": "22319.49"
	}
}`

const testExpectedProxiedResponse = `{"USD":{"ask":22324.79,"bid":22299.15,"last":22319.49}}`

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestProxy(t *testing.T) {
	// Create external http mocks
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", tickerEndpoint, httpmock.NewStringResponder(200, testResponseBody))

	outfile := fmt.Sprintf("/tmp/ticker_proxy_test_%d.json", rand.Int())

	// Remove outfile
	err := os.Remove(outfile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}

	// Create a proxy
	proxy := New(1, "pubkey", "privkey", outfile)
	proxy.SetStream(newTestStream())

	// Fetch data
	err = proxy.Fetch()
	if err != nil {
		t.Fatal(err)
	}

	// Make sure we get the correct response
	if proxy.String() != testExpectedProxiedResponse {
		t.Fatal("Incorrect response body.")
	}

	// Make sure we wrote to outfile
	savedBytes, err := ioutil.ReadFile(outfile)
	if err != nil {
		t.Fatal(err)
	}
	if proxy.String() != string(savedBytes) {
		t.Fatal("Incorrect response body.")
	}

	go proxy.Start()

	// Try the HTTP handler
	rr := httptest.NewRecorder()
	proxy.ServeHTTP(rr, httptest.NewRequest("GET", "http://ticker.openbazaar.org/api", nil))
	result := rr.Result()

	// Check that the response body is correct
	responseBody, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Error(err)
	}

	if string(responseBody) != testExpectedProxiedResponse {
		t.Fatal("Incorrect response body.")
	}

	proxy.Stop()
}

func newTestStream() *health.Stream {
	return health.NewStream()
}
