FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o guardian ./cmd/guardian

FROM alpine:latest

RUN apk add --no-cache ca-certificates git

COPY --from=builder /build/guardian /usr/local/bin/guardian

ENTRYPOINT ["guardian"]
