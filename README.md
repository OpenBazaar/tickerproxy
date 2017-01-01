# TickerProxy

TickerProxy is a simple proxy tool for getting the latest crypto financial data from bitcoinaverage.com. The goal is to provide a caching layer between OpenBazaar nodes and the bitcoinaverage.com infrastructure. Eventually the price logic can be enhanced here, for example by average multiple sources of data.

## Install

```go
go get github.com/OpenBazaar/tickerproxy
```

## Run

```go
go run "$GOPATH/src/github.com/OpenBazaar/tickerproxy/bin/main.go"
```

## Configure

```bash
export TICKER_PROXY_PUBKEY="foo"
export TICKER_PROXY_PRIVKEY="bar"
```