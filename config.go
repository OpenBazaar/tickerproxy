package ticker

import "os"

type Config struct {
	OutPath       string
	AWSS3Region   string
	AWSS3Bucket   string
	BTCAVGPubkey  string
	BTCAVGPrivkey string
	BugsnagAPIKey string
}

func NewConfig() Config {
	return Config{
		OutPath:       getEnvString("TICKER_OUT_PATH", "./"),
		AWSS3Region:   getEnvString("AWS_S3_REGION", ""),
		AWSS3Bucket:   getEnvString("AWS_S3_BUCKET", ""),
		BTCAVGPubkey:  getEnvString("TICKER_BTCAVG_PUBKEY", ""),
		BTCAVGPrivkey: getEnvString("TICKER_BTCAVG_PRIVKEY", ""),
		BugsnagAPIKey: getEnvString("TICKER_BUGSNAG_APIKEY", ""),
	}
}

func getEnvString(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		val = defaultVal
	}
	return val
}
