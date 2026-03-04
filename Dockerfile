# Build stage
FROM golang:1.25.0 as builder

WORKDIR /app

# Cache dependencies separately from source
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o untis-notifier .

# Final stage — minimal image
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/untis-notifier /untis-notifier

ENTRYPOINT ["/untis-notifier"]
