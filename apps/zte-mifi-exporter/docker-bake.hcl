target "docker-metadata-action" {}

variable "APP" {
  default = "zte-mifi-exporter"
}

variable "VERSION" {
  default = "0.6.0"
}

variable "GO_VERSION" {
  // renovate: datasource=docker depName=golang
  default = "1.26.4"
}

group "default" {
  targets = ["image-local"]
}

target "image" {
  inherits = ["docker-metadata-action"]
  args = {
    GO_VERSION = "${GO_VERSION}"
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
