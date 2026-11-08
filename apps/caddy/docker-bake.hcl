target "docker-metadata-action" {}

variable "APP" {
  default = "caddy"
}

variable "VERSION" {
  // renovate: datasource=github-releases depName=caddyserver/caddy
  default = "v2.11.7"
}

variable "SOURCE" {
  default = "https://github.com/caddyserver/caddy"
}

group "default" {
  targets = ["image-local"]
}

target "image" {
  inherits = ["docker-metadata-action"]
  args = {
    VERSION = "${VERSION}"
    // 上游镜像标签不带 v 前缀
    CADDY_VERSION = trimprefix("${VERSION}", "v")
  }
  labels = {
    "org.opencontainers.image.source" = "${SOURCE}"
  }
}

target "image-local" {
  inherits = ["image"]
  output = ["type=docker"]
  tags = ["${APP}:${VERSION}"]
}

target "image-all" {
  inherits = ["image"]
  platforms = [
    "linux/amd64",
    "linux/arm64"
  ]
}
