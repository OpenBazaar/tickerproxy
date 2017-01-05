package tickerproxy

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/gocraft/health"
	"github.com/jarcoal/httpmock"
)

const testResponseBody = `{"BTC": {"low": 0, "high": 1200.00}}`

func TestProxy(t *testing.T) {
	// Create external http mocks
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", tickerEndpoint, httpmock.NewStringResponder(200, testResponseBody))

	// Create a proxy
	proxy := New(1, "pubkey", "privkey")
	proxy.SetStream(newTestStream())

	// Fetch data
	err := proxy.Fetch()
	if err != nil {
		t.Error(err)
	}

	// Make sure we get the correct response
	if proxy.String() != testResponseBody {
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

	if string(responseBody) != testResponseBody {
		t.Fatal("Incorrect response body.")
	}

	proxy.Stop()
}

func newTestStream() *health.Stream {
	return health.NewStream()
}
