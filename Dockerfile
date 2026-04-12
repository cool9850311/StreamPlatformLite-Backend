FROM golang:1.26-alpine

RUN apk add --no-cache ffmpeg su-exec

# Create non-root user before building
RUN adduser -D -u 1001 appuser

WORKDIR /app
COPY . .

WORKDIR /app/Go-Service
RUN go mod download
RUN go build -o main src/main/main.go

# Give appuser ownership of the working directory so it can write application.log
RUN chown -R appuser:appuser /app/Go-Service

RUN chmod +x /app/entrypoint.sh

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["./main"]
