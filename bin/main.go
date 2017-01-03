package main

import (
	"net/http"
	"os"
	"time"

	"strconv"

	"github.com/OpenBazaar/tickerproxy"
)

func main() {
	// Get configuration
	port := getEnvString("TICKER_PROXY_PORT", "8080")
	speed := getEnvString("TICKER_PROXY_SPEED", "10")
	pubkey := getEnvString("TICKER_PROXY_PUBKEY", "")
	privkey := getEnvString("TICKER_PROXY_PRIVKEY", "")

	// Convert speed to an int of seconds, and then into a time.Duration
	speedInt, err := strconv.Atoi(speed)
	if err != nil {
		panic(err)
	}

	tickerDuration := time.Duration(speedInt) * time.Second

	// Create and start a `tickerproxy.Proxy`
	proxy := tickerproxy.New(tickerDuration, pubkey, privkey)
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
