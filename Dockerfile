FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.* .
COPY cmd cmd
COPY pkg pkg
RUN go build -o firstbot FirstBot/cmd/firstbot

FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /app/firstbot .
ENTRYPOINT ["/app/firstbot"]
