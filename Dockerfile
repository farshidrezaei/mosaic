# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/mosaic ./cmd/mosaic

# Runtime Stage with FFmpeg
FROM alpine:3.21

RUN apk add --no-cache ffmpeg tzdata ca-certificates

COPY --from=builder /build/mosaic /usr/local/bin/mosaic

WORKDIR /workspace

ENTRYPOINT ["mosaic"]
CMD ["--help"]
