FROM golang:1.11

WORKDIR $GOPATH/src/github.com/10gen/atlas-service-broker

COPY . .

# Install dependencies
RUN go get -d -v ./...

# Compile and install binary
RUN go install .

# Run binary
CMD ["atlas-service-broker"]
