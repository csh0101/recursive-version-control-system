# Environment Variables
export GIN_MODE=release
export GO111MODULE=on
export CGO_ENABLED=0
export GOPROXY='goproxy.cn,direct'
export GOPRIVATE=""


BUILD_TIME=$(shell date +%FT%T%z)
GIT_REVISION=$(shell git rev-parse --short HEAD)
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_BRANCH=$(shell git rev-parse --symbolic-full-name --abbrev-ref HEAD)
GIT_DIRTY=$(shell test -z "$$(git status --porcelain)" && echo "clean" || echo "dirty")
VERSION=$(shell git describe --tag --abbrev=0 --exact-match HEAD 2>/dev/null || (echo 'Git tag not found, fallback to commit id' >&2; echo ${GIT_REVISION}))

METADATA_PATH=gitlab.hotbot.cc/dcloud/trainer-cloud-disk-api.git/version
INJECT_VARIABLE=-X ${METADATA_PATH}.gitVersion=${VERSION} -X ${METADATA_PATH}.gitCommit=${GIT_COMMIT} -X ${METADATA_PATH}.gitBranch=${GIT_BRANCH} -X ${METADATA_PATH}.gitTreeState=${GIT_DIRTY} -X ${METADATA_PATH}.buildTime=${BUILD_TIME} -X ${METADATA_PATH}.env=${ENV}

FLAGS=-trimpath -ldflags "-s -w ${INJECT_VARIABLE}"

build:
	@echo "Building /cmd/rvcs/main.go..."
	@go build -o cmd/rvcs/rvcs cmd/rvcs/rvcs.go
	@echo "Build completed."
	@echo "cp binary to /Users/d-robotics/lib/go/bin/rvcs"
	@cp cmd/rvcs/rvcs /Users/d-robotics/lib/go/bin/rvcs
