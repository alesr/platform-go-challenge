variable "GO_VERSION" {
  default = "1.24"
}

group "default" {
  targets = ["app", "test"]
}

target "docker-metadata-action" {}

target "build-common" {
  context = "."
  platforms = ["linux/amd64", "linux/arm64"]
}

target "app" {
  inherits = ["build-common"]
  dockerfile = "build/Dockerfile"
  args = {
    GO_VERSION = "${GO_VERSION}"
  }
}

target "test" {
  inherits = ["build-common"]
  dockerfile = "build/Dockerfile.test"
  args = {
    GO_VERSION = "${GO_VERSION}"
  }
}
