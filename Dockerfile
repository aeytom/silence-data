
FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

WORKDIR /go/src/app
COPY . .
RUN go get -d -v
RUN GOOS=linux GOARCH=amd64 go install -ldflags="-w -s"

FROM alpine:latest
WORKDIR /
COPY --from=builder /go/bin/silence-data ./
ENTRYPOINT ["/silence-data"]