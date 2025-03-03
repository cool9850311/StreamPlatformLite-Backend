# Use the official Golang image as the base image
FROM golang:1.22-alpine
RUN apk add --no-cache ffmpeg

WORKDIR /app

COPY . .
WORKDIR /app/Go-Service
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container


# Build the Go app
RUN go build -o main src/main/main.go

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
