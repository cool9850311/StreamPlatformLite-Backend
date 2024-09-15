# StreamPlatformLite-Backend

## Development

### Prerequisites

- Go 1.22 or higher
- Docker
- Docker Compose

### Setup

1. Clone the repository:
    ```sh
    git clone https://github.com/cool9850311/StreamPlatformLite-Backend.git
    cd StreamPlatformLite-Backend
    ```

2. Copy the example environment file and update the environment variables as needed:
    ```sh
    cp .env.example .env
    ```

3. Start the services using Docker Compose:
    ```sh
    cd docker
    docker-compose up -d
    ```

### Running the Application

1. Build and run the application:
    ```sh
    go build -o bin/streamplatformlite-backend Go-Service/src/main.go
    ./bin/streamplatformlite-backend
    ```

### Running Tests

1. Run the tests:
    ```sh
    go test ./...
    ```

### Cleanup

1. Stop the services and clean up:
    ```sh
    cd docker
    docker-compose down
    ```
## contributors
<a href="https://github.com/cool9850311/StreamPlatformLite-Backend/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=cool9850311/StreamPlatformLite-Backend" />
</a>

Made with [contrib.rocks](https://contrib.rocks).
