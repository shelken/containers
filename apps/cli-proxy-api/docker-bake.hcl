target "docker-metadata-action" {}

variable "APP" {
  default = "cli-proxy-api"
}

variable "VERSION" {
  // renovate: datasource=github-releases depName=shelken/CLIProxyAPI
  default = "v6.9.30"
}

variable "SOURCE" {
  default = "https://github.com/shelken/CLIProxyAPI"
}

group "default" {
  targets = ["image-local"]
}

target "image" {
  inherits = ["docker-metadata-action"]
  args = {
    VERSION     = "${VERSION}"
    SOURCE_REPO = "${SOURCE}.git"
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
