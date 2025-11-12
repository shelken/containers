# Build stage
ARG DOCKER_META_VERSION

FROM caddy:${DOCKER_META_VERSION}-builder AS builder

RUN xcaddy build \
  --with github.com/caddy-dns/cloudflare
# --with github.com/mholt/caddy-l4
#    --with github.com/WeidiDeng/caddy-cloudflare-ip

# Final stage
FROM caddy:${DOCKER_META_VERSION}

# Copy the custom-built Caddy binary
COPY --from=builder /usr/bin/caddy /usr/bin/caddy

# Install setcap utility
RUN apk add --no-cache libcap

# Set capabilities to allow binding to privileged ports as non-root user
RUN setcap 'cap_net_bind_service=+ep' /usr/bin/caddy

# Create non-root user
RUN addgroup -g 1000 -S caddy && \
  adduser -u 1000 -S caddy -G caddy

# Ensure directories have proper permissions for both users
RUN chown -R caddy:caddy /etc/caddy /config /data /srv && \
  chmod -R 755 /etc/caddy /config /data /srv

# Keep default behavior (run as root) but allow non-root usage
# Users can override by setting securityContext in Kubernetes
