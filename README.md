### Overview

#### Explorer

The `explorer` binary is a tiny REST API service that provides a simple interface for interacting with the system. It is designed to be lightweight and efficient, making it suitable for high-traffic environments. The `explorer` API is defined in the [internal/explorer/api/routes.go](internal/explorer/api/routes.go) file.

#### Explorerd

The `explorerd` binary is a worker service that performs background tasks and processing. It is designed to be highly scalable and fault-tolerant, making it suitable for large-scale deployments. The `explorerd` service is defined in the [internal/explorerd/main.go](internal/explorerd/main.go) file.

### Running the Project

--------------

#### Using Docker Compose

To run the project using Docker Compose, navigate to the project root directory and run the following command:

```sh
docker-compose -f build/docker/docker-compose.yaml up
```

This will start both the `explorer` and `explorerd` services with all necessary dependencies.

#### Building & Running from Scratch

- Ensure you have you a running instance of the following
  - postgres
  - nats server

```sh
make build-explorer
```

```sh
make build-explorerd
```

### Required Environment Variables

--------------

The following environment variables are required to run the project:

- `EXPLORER_DATABASE_URL`: the URL of the PostgreSQL database
- `NATS_URL`: the URL of the NATS messaging system
- `EXPLORER_TEST_DATABASE_URL` : for running tests
