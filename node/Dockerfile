FROM golang:1.18-alpine as builder

WORKDIR /app

COPY ./node ./node

WORKDIR /app/node

RUN go mod download

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o node ./cmd/node/main.go

################################################################################

FROM alpine

COPY --from=builder /app/node/node /app/node

CMD ["/app/node"]
