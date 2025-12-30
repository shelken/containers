# Build stage
ARG GO_VERSION=1.23.5

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY main.go go.mod ./

RUN go build -o zte-mifi-exporter main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS and tzdata for timezone
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/zte-mifi-exporter .

# Create non-root user
RUN addgroup -g 1000 -S exporter && \
    adduser -u 1000 -S exporter -G exporter

USER exporter

EXPOSE 9586

ENTRYPOINT ["./zte-mifi-exporter"]
