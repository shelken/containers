# Build stage
ARG GO_VERSION=1.25.5

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY main.go go.mod ./

RUN go build -o alertmanager-wechatrobot main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install tzdata for timezone support
RUN apk add --no-cache tzdata

COPY --from=builder /app/alertmanager-wechatrobot .

# Create non-root user
RUN addgroup -g 1000 -S robot &&     adduser -u 1000 -S robot -G robot

USER robot

EXPOSE 8080

ENTRYPOINT ["./alertmanager-wechatrobot"]
