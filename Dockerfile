# build stage
FROM golang:1.19.1 as builder

WORKDIR /go/src/template
# Copy all the Code and stuff to compile everything
COPY go.mod go.sum ./

# Copy all the Code and stuff to compile everything
COPY . .
# Downloads all the dependencies in advance (could be left out, but it's more clear this way)

RUN \
    # Builds the application as a statically linked one, to allow it to run on alpine
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -tags appsec  -o compiled-app ./cmd/httpserver/main

# # Moving the binary to the 'final Image' to make it smaller
FROM alpine:latest

RUN apk add --no-cache libc6-compat
# # `service` should be replaced here as well
COPY --from=builder /go/src/template/compiled-app .

ENV DD_SERVICE="template" \
    DD_TRACE_REDIS_ANALYTICS_ENABLED="true" \
    DD_APPSEC_ENABLED="true" \
    DD_TRACE_ENABLED="true"

CMD ["./compiled-app"]
