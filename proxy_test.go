package ticker

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/gocraft/health"
	"github.com/jarcoal/httpmock"
)

const (
	testFiatResponseBody = `{"BTCUSD": {"ask": "1","bid": "2","last": "3"}}`

	testCryptoResponseBody = `{"BCHBTC": {"ask": "6","bid": "7","last": "8"},"ZECBTC": {"ask": "60","bid": "70","last": "80"}}`

	testCMCResponseBodyPage1 = `{
		"metadata": {"num_cryptocurrencies": 102},
		"data": {
				"1": {
						"symbol": "SOIL",
						"quotes": {"BTC": {"price": 0.0012345}}
				}
		}
	}`

	testCMCResponseBodyPage2 = `{
		"metadata": {"num_cryptocurrencies": 102},
		"data": {
				"101": {
						"symbol": "$$$",
						"quotes": {"BTC": {"price": 0.998877}}
				}
		}
	}`

	testExpectedProxiedResponse = `{"$$$":{"ask":1.0011242625468402,"bid":1.0011242625468402,"last":1.0011242625468402,"type":"crypto"},"BCH":{"ask":0.16666667,"bid":0.14285715,"last":0.125,"type":"crypto"},"BTC":{"ask":1,"bid":1,"last":1,"type":"crypto"},"SOIL":{"ask":810.0445524503848,"bid":810.0445524503848,"last":810.0445524503848,"type":"crypto"},"USD":{"ask":1,"bid":2,"last":3,"type":"fiat"},"ZEC":{"ask":0.016666668,"bid":0.014285714,"last":0.0125,"type":"crypto"}}`
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestProxy(t *testing.T) {
	// Create external http mocks
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", btcavgFiatEndpoint, httpmock.NewStringResponder(200, testFiatResponseBody))
	httpmock.RegisterResponder("GET", btcavgCryptoEndpoint, httpmock.NewStringResponder(200, testCryptoResponseBody))
	httpmock.RegisterResponder("GET", cmcBaseEndpoint+"1", httpmock.NewStringResponder(200, testCMCResponseBodyPage1))
	httpmock.RegisterResponder("GET", cmcBaseEndpoint+"101", httpmock.NewStringResponder(200, testCMCResponseBodyPage2))

	// Prepare outfile
	outfile := fmt.Sprintf("/tmp/ticker_proxy_test_%d.json", rand.Int())
	err := os.Remove(outfile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}

	// Create a proxy with a test output stream
	proxy, err := New(1, "pubkey", "privkey", outfile, TestS3Region, TestS3Region)
	if err != nil {
		t.Fatal(err)
	}
	proxy.SetStream(health.NewStream())

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
}
