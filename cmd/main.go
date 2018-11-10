package main

import (
	"log"
	"os"

	"github.com/OpenBazaar/tickerproxy"
	"github.com/gocraft/health"
	"github.com/gocraft/health/sinks/bugsnag"
)

func main() {
	conf := ticker.NewConfig()

	writers, err := getWriters(conf.OutPath, conf.AWSS3Region, conf.AWSS3Bucket)
	if err != nil {
		log.Fatalln("creating writers failed:", err)
	}

	err = ticker.Fetch(
		newHealthStream(conf.BugsnagAPIKey),
		conf.BTCAVGPubkey,
		conf.BTCAVGPrivkey,
		writers...)
	if err != nil {
		log.Fatalln("ticker failed:", err)
	}
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
