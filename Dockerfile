# FROM golang:1.21

# WORKDIR /app

# COPY go.mod go.sum ./

# RUN go mod download

# COPY . ./

# RUN CGO_ENABLED=0 GOOS=linux go build -o /main

# CMD ["/main"] 

# Stage 1: Build the application
FROM golang:1.21 AS BUILDER

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

# Stage 2: Create a minimal image with the built binary
FROM alpine:latest AS RUNNER

COPY --from=builder /main /

CMD ["/main"]
