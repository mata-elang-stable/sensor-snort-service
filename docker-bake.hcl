variable "APP_VERSION" {
    default = "latest"
}

variable "APP_COMMIT_SHORTHASH" {
    default = "latest"
}

variable "DOCKER_IMAGE_TAG" {
    default = "2"
}

variable "DOCKER_IMAGE_NAME" {
    default = "mfscy/snort3-parser"
}

group "default" {
    targets = ["build"]
}

target "build" {
    context = "./"
    dockerfile = "./Dockerfile"
    args = {
        APP_VERSION = APP_VERSION
        APP_COMMIT = APP_COMMIT_SHORTHASH
    }
    annotations = [
        "org.opencontainers.image.source=https://github.com/fadhilyori/sensor-parser",
        "org.opencontainers.image.description=Parse sensor snort data to Mata Elang Defense Center",
        "org.opencontainers.image.version=${APP_VERSION}",
        "org.opencontainers.image.revision=${APP_COMMIT_SHORTHASH}",
        "org.opencontainers.image.authors=Fadhil Yori Hibatullah",
        "org.opencontainers.image.license=MIT",
    ]
    platforms = ["linux/amd64", "linux/arm64"]
    tags = ["${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"]
}