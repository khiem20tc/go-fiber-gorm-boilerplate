# BOTP Gateway

![Golang](https://img.shields.io/badge/language-Golang-blue.svg)

The BOTP Gateway is a GoLang-based project designed to serve as a versatile gateway for various API management

## Features

- **Load Balancing:** Distribute incoming requests evenly across multiple backend services to ensure optimal resource utilization and fault tolerance.

- **Rate Limiting:** Protect services from excessive traffic by setting rate limits on incoming requests.

- **API Forwarding:** Easily configure and manage the routing of API requests from clients to the appropriate backend services.

## Getting Started

### Prerequisites

Before you can run the BOTP Gateway, ensure you have the following prerequisites installed:

- [Go](https://golang.org/dl/): Make sure you have Go installed on your system.

### Installation

1. Clone the repository to your local machine:
```bash
git clone git@github.com:B-K-Labs/BOTP-Gateway.git
```

2. Create a `.env` file and define the necessary environment variables. These variables will be used for configuration. Example:

```
DATABASE_URL="localhost:5432"
```

3. Run the BOTP Gateway using the following command:
```bash
go run main.go
```

4. Gen swagger API document using the following command:
```bash
go run ./scripts/gen-swagger/main.go
```

## Usage

Once the BOTP Gateway is up and running, you can start sending API requests to it, and it will handle load balancing, rate limiting, and forwarding based on the configuration.