FROM golang:1.10
WORKDIR /go/src/github.com/OpenBazaar/tickerproxy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags '-extldflags "-static"' -o /opt/tickerfetcher ./fetch/*.go

FROM scratch
WORKDIR /var/lib/ticker
COPY --from=0 /opt/tickerfetcher /opt/tickerfetcher
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
CMD ["/opt/tickerfetcher"]
