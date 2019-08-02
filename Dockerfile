# Build stage uses a full golang image to build a statically linked binary
FROM golang:1.12 AS builder
WORKDIR /usr/src
COPY . .

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/atlas-service-broker

# Run stage uses a much smaller base image to run the prebuilt binary
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Copy binary from build stage
WORKDIR /root
COPY --from=builder /usr/src/bin .

CMD ["./atlas-service-broker"]