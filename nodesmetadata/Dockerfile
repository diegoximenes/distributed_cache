FROM golang:1.18-alpine as builder

WORKDIR /app

COPY ./nodesmetadata ./nodesmetadata
COPY ./util ./util

WORKDIR /app/nodesmetadata

RUN go mod download

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o nodesmetadata ./cmd/nodesmetadata/main.go

################################################################################

FROM alpine

COPY --from=builder /app/nodesmetadata/nodesmetadata /app/nodesmetadata

CMD ["/app/nodesmetadata"]
