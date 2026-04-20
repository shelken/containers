FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

ARG VERSION=dev
ARG SOURCE_REPO=https://github.com/shelken/CLIProxyAPI.git
ARG COMMIT=none
ARG BUILD_DATE=unknown

RUN git clone --depth=1 --branch "${VERSION}" "${SOURCE_REPO}" .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.Commit=${COMMIT}' -X 'main.BuildDate=${BUILD_DATE}'" -o ./CLIProxyAPI ./cmd/server/

FROM alpine:3.22.0

RUN apk add --no-cache tzdata

RUN mkdir /CLIProxyAPI

COPY --from=builder /src/CLIProxyAPI /CLIProxyAPI/CLIProxyAPI
COPY --from=builder /src/config.example.yaml /CLIProxyAPI/config.example.yaml

WORKDIR /CLIProxyAPI

EXPOSE 8317

ENV TZ=Asia/Shanghai

RUN cp /usr/share/zoneinfo/${TZ} /etc/localtime && echo "${TZ}" > /etc/timezone

CMD ["./CLIProxyAPI"]
