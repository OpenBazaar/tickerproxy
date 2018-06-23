package ticker

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"

	"github.com/gocraft/health"
	"github.com/jarcoal/httpmock"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestFetch(t *testing.T) {
	stream := health.NewStream()
	stream.AddSink(&health.WriterSink{os.Stdout})

	disableMocksFn := createHTTPMocks()
	defer disableMocksFn()

	// Prepare outfiles path
	outfilePath := fmt.Sprintf("/tmp/ticker_proxy_test_%d", rand.Int())
	err := os.Remove(outfilePath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	err = os.Mkdir(outfilePath, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	// Fetch data
	Fetch(stream, "pubkey", "privkey", func(_ *health.Job, data []byte) error {
		if string(data) != testExpectedFetchData {
			t.Fatal("Fetch returned incorrect data\nGot:", string(data), "\nWanted:", testExpectedFetchData)
		}
		return nil
	}, NewFileSystemWriter(outfilePath))

	// Make sure we wrote to outfiles
	savedBytes, err := ioutil.ReadFile(path.Join(outfilePath, "rates"))
	if err != nil {
		t.Fatal(err)
	}
	if string(savedBytes) != testExpectedFetchData {
		t.Fatal("Incorrect rates outfile contents:", string(savedBytes))
	}
	savedBytes, err = ioutil.ReadFile(path.Join(outfilePath, "whitelist"))
	if err != nil {
		t.Fatal(err)
	}
	if string(savedBytes) != string(PinnedSymbolsToIDsJSON()) {
		t.Fatal("Incorrect whitelist outfile contents:", string(savedBytes))
	}
}

func createHTTPMocks() func() {
	httpmock.Activate()
	for endpoint, resp := range httpMocks {
		httpmock.RegisterResponder("GET", endpoint, httpmock.NewStringResponder(200, resp))
	}
	return httpmock.DeactivateAndReset
}
