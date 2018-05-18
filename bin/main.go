package main

import (
	"log"
	"os"

	"github.com/OpenBazaar/tickerproxy"
	"github.com/gocraft/health"
	"github.com/gocraft/health/sinks/bugsnag"
)

type config struct {
	outPath       string
	awsS3Region   string
	awsS3Bucket   string
	btcAVGPubkey  string
	btcAVGPrivkey string
	bugsnagAPIKey string
}

func newConfig() config {
	return config{
		outPath:       getEnvString("TICKER_OUT_PATH", "./"),
		awsS3Region:   getEnvString("AWS_S3_REGION", ""),
		awsS3Bucket:   getEnvString("AWS_S3_BUCKET", ""),
		btcAVGPubkey:  getEnvString("TICKER_BTCAVG_PUBKEY", ""),
		btcAVGPrivkey: getEnvString("TICKER_BTCAVG_PRIVKEY", ""),
		bugsnagAPIKey: getEnvString("TICKER_BUGSNAG_APIKEY", ""),
	}
}

func main() {
	conf := newConfig()

	writers, err := getWriters(conf.outPath, conf.awsS3Region, conf.awsS3Bucket)
	if err != nil {
		log.Fatalln("creating writers failed:", err)
	}

	err = ticker.Fetch(
		newHealthStream(conf.bugsnagAPIKey),
		conf.btcAVGPubkey,
		conf.btcAVGPrivkey,
		writers...)
	if err != nil {
		log.Fatalln("ticker failed:", err)
	}
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

func getWriters(outfile string, s3Region string, s3Bucket string) ([]ticker.Writer, error) {
	writers := []ticker.Writer{}

	if outfile != "" {
		writers = append(writers, ticker.NewFileSystemWriter(outfile))
	}

	if s3Region != "" {
		writer, err := ticker.NewS3Writer(s3Region, s3Bucket)
		if err != nil {
			return nil, err
		}
		writers = append(writers, writer)
	}

	return writers, nil
}
