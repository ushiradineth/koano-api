# https://github.com/GoogleCloudPlatform/golang-samples/blob/main/run/helloworld/Dockerfile

# Use the offical golang image to create a binary.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.22 AS builder

# Create and change to the app directory.
WORKDIR /app

# Copy local code to the container image.
COPY . ./

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
RUN go mod download

# Build the binary.
RUN go build -v -o server

# Run the tests in the container
RUN go test -v -cover ./...

# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM debian:bookworm-slim AS deploy

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
  ca-certificates && \
  rm -rf /var/lib/apt/lists/*

WORKDIR /

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /app/server

# Run the web service on container startup.
CMD ["/app/server"]