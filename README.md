# TickerProxy

TickerProxy gathers the latest financial data from bitcoinaverage.com. The goal is to provide a caching layer between OpenBazaar nodes and the bitcoinaverage.com infrastructure. It provides exchange rates against BTC for all known fiat symbols and a few crypto symbols.

It can writes responses to a local file and/or AWS S3.

Get your account's API public and private keys from bitcoinaverage.com.

## Install

```go
go get github.com/OpenBazaar/tickerproxy
```

## Run

```go
go run "$GOPATH/src/github.com/OpenBazaar/tickerproxy/cmd/main.go"
```

## Configuration and defaults

```bash
export TICKER_PROXY_SPEED="10"              # Number of seconds to wait between updates
export TICKER_PROXY_PUBKEY=""               # API public key from bitcoinaverage.com
export TICKER_PROXY_PRIVKEY=""              # API private key from bitcoinaverage.com
export TICKER_PROXY_OUTFILE="/path/to/file" # A file to write outputs to
export AWS_REGION="us-east-1"               # An AWS region to write to
export AWS_S3_BUCKET="openbazaar-ticker"    # An AWS bucket to write outputs to
export TICKER_BUGSNAG_APIKEY="secretkey"    # A Bugsnag key for error monitoring
```