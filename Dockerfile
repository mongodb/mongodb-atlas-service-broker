# Build stage uses a full golang image to build a statically linked binary
FROM golang:1.12 AS builder
WORKDIR /usr/src

# Download and cache dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN ./dev/scripts/build-production-binary.sh bin/atlas-service-broker

# Run stage uses a much smaller base image to run the prebuilt binary
FROM alpine:3.10
RUN apk --no-cache add ca-certificates

# Copy binary from build stage
WORKDIR /root
COPY --from=builder /usr/src/bin .

CMD ["./atlas-service-broker"]
