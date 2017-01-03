package tickerproxy

import (
	"testing"

	"github.com/jarcoal/httpmock"
)

const testResponseBody = `{"BTC": {"low": 0, "high": 1200.00}}`

func TestProxyFetch(t *testing.T) {
	// Create external http mocks
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", tickerEndpoint, httpmock.NewStringResponder(200, testResponseBody))

	// Create a proxy
	proxy := New(1, "pubkey", "privkey")

	// Fetch data
	err := proxy.Fetch()
	if err != nil {
		t.Error(err)
	}

	if string(proxy.lastResponseBody) != testResponseBody {
		t.Fatal("Incorrect response body.")
	}
}
