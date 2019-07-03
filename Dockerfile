FROM golang:1.11

RUN mkdir -p /usr/src
WORKDIR /usr/src

COPY . .

RUN mkdir bin

# Compile and install binary
RUN go build -o bin/atlas-service-broker

# Run binary
CMD ["bin/atlas-service-broker"]
