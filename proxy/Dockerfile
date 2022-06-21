FROM golang:1.18-alpine as builder

WORKDIR /app

COPY ./proxy ./proxy
COPY ./util ./util

WORKDIR /app/proxy

RUN go mod download

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o proxy ./cmd/proxy/main.go

################################################################################

FROM alpine

COPY --from=builder /app/proxy/proxy /app/proxy

CMD ["/app/proxy"]
