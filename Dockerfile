FROM golang:1.10
WORKDIR /go/src/github.com/OpenBazaar/tickerproxy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags '-extldflags "-static"' -o /opt/ticker ./bin/*.go

FROM scratch
WORKDIR /var/lib/ticker
COPY --from=0 /opt/ticker /opt/ticker
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
CMD ["/opt/ticker"]
