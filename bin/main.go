package main

import (
	"net/http"
	"os"

	"github.com/OpenBazaar/tickerproxy"
)

func main() {
	port := getEnvString("TICKER_PROXY_PORT", "8080")
	pubkey := getEnvString("TICKER_PROXY_PUBKEY", "")
	privkey := getEnvString("TICKER_PROXY_PRIVKEY", "")

	proxy := tickerproxy.New(pubkey, privkey)
	go proxy.Start()

	http.ListenAndServe(":"+port, proxy)
}

func getEnvString(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		val = defaultVal
	}
	return val
}
