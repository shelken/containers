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

FROM builder AS browser-assets

ENV HOME=/tmp/camoufox-home

RUN mkdir -p "$HOME"

RUN camoufox fetch

# 变更原因：AIstudioProxyAPI 在 Linux 容器内固定使用 Linux Camoufox 指纹，最终镜像不需要 macOS/Windows 字体。
RUN rm -rf \
      "$HOME/.cache/camoufox/fonts/macos" \
      "$HOME/.cache/camoufox/fonts/windows"

FROM builder AS app-assets

WORKDIR /src

RUN python scripts/update_browserforge_data.py

# 变更原因：生产 headless 服务只需要 API、浏览器控制和流式代理代码，不需要仓库元数据、文档、测试、GUI 和前端源码。
RUN rm -rf \
      .git \
      .github \
      docker \
      docs \
      tests \
      gui \
      scripts \
      static \
      .dockerignore \
      .env.example \
      .gitattributes \
      .gitignore \
      CONTRIBUTING.md \
      LICENSE \
      README.md \
      monkeytype_config.py \
      pyrightconfig.json \
      simple_launcher.py \
      simple_start.sh \
      supervisord.conf \
      支持作者.jpg \
      poetry.lock \
      pyproject.toml

# 变更原因：GUI 和本地开发工具已删除，相关 Python 包不会被生产 API 路径导入。
RUN rm -rf \
      /opt/venv/lib/python3.10/site-packages/customtkinter \
      /opt/venv/lib/python3.10/site-packages/customtkinter-*dist-info \
      /opt/venv/lib/python3.10/site-packages/PIL \
      /opt/venv/lib/python3.10/site-packages/pillow.libs \
      /opt/venv/lib/python3.10/site-packages/pillow-*dist-info \
      /opt/venv/lib/python3.10/site-packages/flask \
      /opt/venv/lib/python3.10/site-packages/flask-*dist-info \
      /opt/venv/lib/python3.10/site-packages/werkzeug \
      /opt/venv/lib/python3.10/site-packages/werkzeug-*dist-info \
      /opt/venv/lib/python3.10/site-packages/jinja2 \
      /opt/venv/lib/python3.10/site-packages/jinja2-*dist-info \
      /opt/venv/lib/python3.10/site-packages/itsdangerous \
      /opt/venv/lib/python3.10/site-packages/itsdangerous-*dist-info

# 变更原因：runtime 不执行安装与测试，删除 Python 缓存和测试目录可减少最终复制内容。
RUN find /opt/venv -type d \( -name '__pycache__' -o -name 'tests' -o -name 'test' \) -prune -exec rm -rf '{}' + \
    && find /opt/venv -type f \( -name '*.pyc' -o -name '*.pyo' \) -delete

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

COPY --from=app-assets /opt/venv /opt/venv
COPY --from=app-assets /src /app
COPY --from=browser-assets /tmp/camoufox-home/.cache/camoufox /home/app/.cache/camoufox

RUN mkdir -p \
      /app/auth_profiles/active \
      /app/auth_profiles/saved \
      /app/auth_profiles/emergency \
      /app/logs \
      /app/.cache \
      /home/app/.mozilla \
    && chown -R app:app /app /home/app /opt/venv

USER 1000:1000

EXPOSE 2048 3120

CMD ["python", "launch_camoufox.py", "--headless", "--server-port", "2048", "--stream-port", "3120", "--helper", ""]
