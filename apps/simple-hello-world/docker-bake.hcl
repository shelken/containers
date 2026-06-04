target "docker-metadata-action" {}

variable "APP" {
  default = "simple-hello-world"
}

variable "VERSION" {
  default = "0.1.0"
}

variable "SOURCE" {
  default = "https://github.com/shelken/containers"
}

group "default" {
  targets = ["image-local"]
}

target "image" {
  inherits = ["docker-metadata-action"]
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
