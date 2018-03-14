package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/OpenBazaar/tickerproxy"
	"github.com/gocraft/health"
	"github.com/gocraft/health/sinks/bugsnag"
)

func main() {
	// Get configuration
	speed := getEnvString("TICKER_PROXY_SPEED", "900")
	pubkey := getEnvString("TICKER_PROXY_PUBKEY", "")
	privkey := getEnvString("TICKER_PROXY_PRIVKEY", "")
	outfile := getEnvString("TICKER_PROXY_OUTFILE", "/var/lib/tickerproxy/ticker_data.json")
	bugsnagAPIKey := getEnvString("TICKER_BUGSNAG_APIKEY", "")
	awsRegion := getEnvString("AWS_REGION", "")
	s3Bucket := getEnvString("AWS_S3_BUCKET", "")

	// Create instrumentation stream
	stream := newHealthStream(bugsnagAPIKey)

	// Convert speed to an int of seconds, and then into a time.Duration
	speedInt, err := strconv.Atoi(speed)
	if err != nil {
		panic(err)
	}

	// Create and start a `tickerproxy.Proxy`
	proxy, err := tickerproxy.New(speedInt, pubkey, privkey, outfile, awsRegion, s3Bucket)
	if err != nil {
		fmt.Printf("ticker failed: %s", err)
	}

	proxy.SetStream(stream)
	go proxy.Start()
	stream.Event("started")

	// Wait for shutdown signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	<-c
	stream.Event("shutdown")
	proxy.Stop()
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
