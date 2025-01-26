FROM golang:1.22.10-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o shortener ./cmd/shortener



FROM scratch

COPY --from=builder /app/shortener /shortener
ENTRYPOINT ["/shortener"]
