package main

import (
	"os"

	ticker "github.com/OpenBazaar/tickerproxy"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gocraft/health"
)

func main() {
	lambda.Start(Fetch)
}

func Fetch() {
	conf := ticker.NewConfig()

	stream := health.NewStream()
	stream.AddSink(&health.WriterSink{Writer: os.Stdout})

	kvs := map[string]string{
		"region":       conf.AWSS3Region,
		"bucket":       conf.AWSS3Bucket,
		"btcAvgPubKey": conf.BTCAVGPubkey,
	}

	writer, err := ticker.NewS3Writer(conf.AWSS3Region, conf.AWSS3Bucket)
	if err != nil {
		stream.EventErrKv("new_s3_writer", err, kvs)
		os.Exit(1)
	}

	err = ticker.Fetch(stream, conf, writer)
	if err != nil {
		stream.EventErrKv("new_s3_writer", err, kvs)
		os.Exit(1)
	}
}

func getEnvString(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		val = defaultVal
	}
	return val
}
