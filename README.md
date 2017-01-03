# TickerProxy

TickerProxy is a simple proxy tool for getting the latest crypto financial data from bitcoinaverage.com. The goal is to provide a caching layer between OpenBazaar nodes and the bitcoinaverage.com infrastructure. Eventually the price logic can be enhanced here, for example by average multiple sources of data.

Get your account's API public and private keys from bitcoinaverage.com.

## Install

```go
go get github.com/OpenBazaar/tickerproxy
```

## Run

```go
go run "$GOPATH/src/github.com/OpenBazaar/tickerproxy/bin/main.go"
```

## Configuration and defaults

```bash
export TICKER_PROXY_PORT="8080" # Port to listen on
export TICKER_PROXY_SPEED="10"  # Number of seconds to wait between updates
export TICKER_PROXY_PUBKEY=""   # API public key from bitcoinaverage.com
export TICKER_PROXY_PRIVKEY=""  # API private key from bitcoinaverage.com
```