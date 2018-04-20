APPLICATION_NAME = discord1111resolver
VERSION = 0.7.0
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
COMMIT = $(shell git rev-parse HEAD)

LD_FLAGS = -X "main.applicationName=${APPLICATION_NAME}" -X "main.version=${VERSION}" -X "main.branch=${BRANCH}" -X "main.commit=${COMMIT}"

# builds and formats the project with the built-in Golang tool
build:
	@go build -ldflags '${LD_FLAGS}' ./cmd/discord1111resolver

# installs and formats the project with the built-in Golang tool
install:
	@go install -ldflags '${LD_FLAGS}' ./cmd/discord1111resolver
