# Use an official Go runtime as a parent image
FROM golang:1.22.5-alpine

# Set the Current Working Directory inside the container
WORKDIR /app

# Install Air for hot-reloading
RUN apk update && apk add --no-cache git bash curl && \
    curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Install wait-for-it script
COPY wait-for-it.sh /bin/wait-for-it.sh
RUN chmod +x /bin/wait-for-it.sh

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run Air for hot-reloading
CMD ["wait-for-it.sh", "chromium", "3000", "--", "air"]
