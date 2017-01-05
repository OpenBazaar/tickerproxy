package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/OpenBazaar/tickerproxy"
	"github.com/gocraft/health"
	"github.com/gocraft/health/sinks/bugsnag"
)

func main() {
	// Get configuration
	port := getEnvString("TICKER_PROXY_PORT", "8080")
	speed := getEnvString("TICKER_PROXY_SPEED", "10")
	pubkey := getEnvString("TICKER_PROXY_PUBKEY", "")
	privkey := getEnvString("TICKER_PROXY_PRIVKEY", "")
	bugsnagAPIKey := getEnvString("TICKER_BUGSNAG_APIKEY", "")

	// Create instrumentation stream
	stream := newHealthStream(bugsnagAPIKey)

	// Convert speed to an int of seconds, and then into a time.Duration
	speedInt, err := strconv.Atoi(speed)
	if err != nil {
		panic(err)
	}

	// Create and start a `tickerproxy.Proxy`
	proxy := tickerproxy.New(speedInt, pubkey, privkey)
	proxy.SetStream(stream)
	go proxy.Start()

	// Listen for http requests
	http.ListenAndServe(":"+port, proxy)
}

func getEnvString(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		val = defaultVal
	}
	return val
}

func newHealthStream(bugsnagAPIKey string) *health.Stream {
	stream := health.NewStream()
	stream.AddSink(&health.WriterSink{os.Stdout})

	if bugsnagAPIKey != "" {
		stream.AddSink(bugsnag.NewSink(&bugsnag.Config{APIKey: bugsnagAPIKey}))
	}

	return stream
}
