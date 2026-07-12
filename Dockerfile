# --- Build stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Cache dependencies first (go.sum is optional until we add a driver).
COPY go.* ./
RUN go mod download

# Build the statically linked binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /barterswap .

# --- Runtime stage ---
FROM alpine:3.20

# Non-root user for the runtime.
RUN adduser -D -H app
USER app

COPY --from=builder /barterswap /barterswap

EXPOSE 8080
ENTRYPOINT ["/barterswap"]
