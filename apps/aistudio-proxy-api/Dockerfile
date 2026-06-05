# syntax=docker/dockerfile:1.7
FROM python:3.10-slim-bookworm AS builder

ARG DEBIAN_FRONTEND=noninteractive
ARG VERSION=v4.1.2_py
ARG SOURCE_REPO=https://github.com/CJackHwang/AIstudioProxyAPI.git
ARG POETRY_VERSION=1.8.3

ENV POETRY_HOME=/opt/poetry \
    POETRY_NO_INTERACTION=1 \
    POETRY_VIRTUALENVS_CREATE=false \
    VIRTUAL_ENV=/opt/venv \
    PATH=/opt/poetry/bin:/opt/venv/bin:$PATH

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl git \
    && rm -rf /var/lib/apt/lists/*

RUN python -m venv "$VIRTUAL_ENV" \
    && curl -sSL https://install.python-poetry.org | python3 - --version "$POETRY_VERSION"

WORKDIR /src
RUN git clone --depth=1 --branch "$VERSION" "$SOURCE_REPO" .

RUN poetry install --only main --no-root --no-ansi

FROM python:3.10-slim-bookworm

ARG DEBIAN_FRONTEND=noninteractive

ENV VIRTUAL_ENV=/opt/venv \
    PATH=/opt/venv/bin:$PATH \
    PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    HOME=/home/app \
    PLAYWRIGHT_BROWSERS_PATH=/home/app/.cache/ms-playwright \
    SERVER_PORT=2048 \
    DEFAULT_FASTAPI_PORT=2048 \
    DEFAULT_CAMOUFOX_PORT=9222 \
    STREAM_PORT=3120 \
    SERVER_LOG_LEVEL=INFO \
    DEBUG_LOGS_ENABLED=false \
    TRACE_LOGS_ENABLED=false \
    AUTO_SAVE_AUTH=false \
    AUTO_AUTH_ROTATION_ON_STARTUP=false \
    COOKIE_REFRESH_ENABLED=true \
    COOKIE_REFRESH_ON_REQUEST_ENABLED=true \
    COOKIE_REFRESH_ON_SHUTDOWN=true \
    INTERNAL_CAMOUFOX_PROXY="" \
    GUI_DEFAULT_HELPER_ENDPOINT=""

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
      ca-certificates \
      fonts-liberation \
      libasound2 \
      libatk-bridge2.0-0 \
      libatk1.0-0 \
      libcups2 \
      libdbus-1-3 \
      libdrm2 \
      libgbm1 \
      libgtk-3-0 \
      libnspr4 \
      libnss3 \
      libpango-1.0-0 \
      libpangocairo-1.0-0 \
      libu2f-udev \
      libx11-6 \
      libx11-xcb1 \
      libxcb1 \
      libxcomposite1 \
      libxdamage1 \
      libxext6 \
      libxfixes3 \
      libxrandr2 \
      libxrender1 \
      libxtst6 \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd --gid 1000 app \
    && useradd --uid 1000 --gid app --home-dir /home/app --create-home --shell /usr/sbin/nologin app

WORKDIR /app

COPY --from=builder /opt/venv /opt/venv
COPY --from=builder /src /app

RUN mkdir -p \
      /app/auth_profiles/active \
      /app/auth_profiles/saved \
      /app/auth_profiles/emergency \
      /app/logs \
      /app/.cache \
      /home/app/.mozilla \
    && camoufox fetch \
    && python scripts/update_browserforge_data.py \
    && chown -R app:app /app /home/app /opt/venv

USER 1000:1000

EXPOSE 2048 3120

CMD ["python", "launch_camoufox.py", "--headless", "--server-port", "2048", "--stream-port", "3120", "--helper", ""]
