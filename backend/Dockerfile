##
## Build onchain-nonmerklized-issuer-demo
##
FROM golang:1.20-bullseye as base

WORKDIR /build

COPY . .
RUN go mod download

RUN go build -o ./onchain main.go

# Build an onchain-nonmerklized-issuer-demo image
FROM alpine:3.18.0

RUN apk add --no-cache libstdc++ gcompat libgomp

COPY onchain-issuer.settings.yaml /app/onchain-issuer.settings.yaml
COPY resolvers.settings.yaml /app/resolvers.settings.yaml
COPY keys/* /app/keys/

COPY --from=base /build/onchain /app/onchain
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=base "/go/pkg/mod/github.com/iden3/wasmer-go@v0.0.1/wasmer/packaged/" \
  "/go/pkg/mod/github.com/iden3/wasmer-go@v0.0.1/wasmer/packaged/"


WORKDIR /app

ENV HOST=0.0.0.0
ENV PORT=3333

# Command to run
ENTRYPOINT ["/app/onchain"]

