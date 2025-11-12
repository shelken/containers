# Build stage
ARG DOCKER_META_VERSION

FROM caddy:${DOCKER_META_VERSION}-builder AS builder

RUN xcaddy build \
  --with github.com/caddy-dns/cloudflare
# --with github.com/mholt/caddy-l4
#    --with github.com/WeidiDeng/caddy-cloudflare-ip

# support non root
RUN setcap cap_net_bind_service=+eip /usr/bin/caddy

# Final stage
FROM caddy:${DOCKER_META_VERSION}

# Copy the custom-built Caddy binary
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
