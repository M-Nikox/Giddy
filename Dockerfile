# Stage 1: build a static Go binary
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server main.go

# Stage 2: tiny runtime image
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server .
COPY index.html .
EXPOSE 8080
CMD ["./server"]
