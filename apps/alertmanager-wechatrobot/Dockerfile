# Build stage
ARG GO_VERSION

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY main.go go.mod ./

RUN go build -o alertmanager-wechatrobot main.go

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/alertmanager-wechatrobot .

# Create non-root user
RUN addgroup -g 1000 -S robot &&     adduser -u 1000 -S robot -G robot

USER robot

EXPOSE 8080

ENTRYPOINT ["./alertmanager-wechatrobot"]
