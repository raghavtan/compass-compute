# compass-compute CLI

A command-line interface (CLI) tool to manage compass components.

## Overview

`compass-compute` is a CLI tool designed to simplify the management and processing of various components within the compass ecosystem.

## Prerequisites

- Go (version 1.18 or later)
- Docker (latest version recommended)
- Make

## Usage

```bash
./compass-compute <component-name>
```

Replace `<component-name>` with the actual name of the component you want to process.

## Building the CLI

To build the CLI tool from the source code:

1.  **Clone the repository (if you haven't already):**
    ```bash
    git clone <repository-url>
    cd compass-compute
    ```

2.  **Build the binary:**
    ```bash
    make build
    ```
    This will create an executable file named `compass-compute` in the project's root directory.

## Running Tests

To run the automated tests:

```bash
make test
```

## Linting

To check the code for style and potential errors using `golangci-lint`:

```bash
make lint
```
It's recommended to run the linter before committing any changes.

## Docker

### Building the Docker Image

To build a Docker image for the CLI:

```bash
make docker-build
```
This will use the `build/Dockerfile` to create a container image.

### Running with Docker

Once the image is built (e.g., `compass-compute:latest`), you can run the CLI using Docker:

```bash
docker run compass-compute:latest <component-name>
```

## Local Development Setup

To set up your local development environment, including installing necessary Go tools and linters:

```bash
make setup
```

## Makefile Targets

The `Makefile` provides several useful targets:

-   `setup`: Installs Go tools and linters.
-   `build`: Compiles the Go application.
-   `test`: Runs Go tests.
-   `lint`: Runs `golangci-lint`.
-   `docker-build`: Builds the Docker image using `build/Dockerfile`.
-   `clean`: Removes build artifacts and the compiled binary.
-   `all`: Runs `lint`, `test`, and `build`.

## Contributing

[Details on how to contribute to this project - to be added]
