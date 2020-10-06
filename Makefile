ROOT_DIR        := $(abspath $(dir $(lastword ${MAKEFILE_LIST})))
BUILD_DIR       := ${ROOT_DIR}/build
BINARY_DIR      := ${BUILD_DIR}/bin

TARGET          := faceit-users
MODULE          := github.com/pavelmemory/${TARGET}

VERSION         := $(shell git describe --tags 2>/dev/null | echo 'N/A')
BUILD_TIME      := $(shell date +"%Y%m%d.%H%M%S")
COMMIT_SHA      := $(shell git rev-parse --short HEAD)

GO_LDFLAGS      := -ldflags '-X ${MODULE}/internal.CommitSHA=${COMMIT_SHA} \
                             -X ${MODULE}/internal.BuildTimestamp=${BUILD_TIME} \
                             -X ${MODULE}/internal.Version=${VERSION}'

# enables verbose mode for task execution, add `V=on` variable to enable it like in example below
# V=on make clean
ifeq ($(V),on)
	Q =
else
	Q = @
endif

help: ## Lists all commands.
	${Q} fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: clean-go ## Builds executable binary of the service
	${Q} mkdir -p ${BINARY_DIR}
	${Q} go build ${GO_LDFLAGS} -o ${BINARY_DIR}/${TARGET} ${MODULE}/cmd

.PHONY: clean
clean: clean-go ## Removes all building artifacts to start build process from scratch
	${Q} rm -rf ${BUILD_DIR}

.PHONY: clean-go
clean-go: clean-go-build clean-go-test ## Removes temporary files, build artifacts and caches of the Go sources compilation

.PHONY: clean-go-build
clean-go-build: ## Removes temporary files generated after build of Go source code
	${Q} go clean -cache

.PHONY: clean-go-test
clean-go-test: ## Removes temporary files generated after build of Go tests and a cached test results
	${Q} go clean -testcache

.PHONY: test
test: clean-go-test ## Runs unit tests for all packages
	${Q} go test ${MODULE}/...

.PHONY: integration-env-up
integration-env-up: ## Spins up testing environment locally
	${Q} docker-compose -f docker-compose.test.yaml up --detach

.PHONY: integration-env-ready
integration-env-ready: ## Checks if service is ready for use
	${Q} [ `curl -s -o /dev/null -w "%{http_code}" localhost:8080/-/readiness` -eq 200 ] && { echo serice is ready for use!; } || { echo serice is not yet ready, please wait...; }

.PHONY: integration-env-down
integration-env-down: ## Cleans up local testing environment
	${Q} docker-compose -f docker-compose.test.yaml rm -f

.PHONY: integration-test
integration-test: ## Runs integration tests on local environment
	${Q} go test -count=1 -tags integration ./...

.PHONY: format
format: install-goimports ## Formats Go source code according to the unified code-style
	${Q} goimports -local ${MODULE} -w $(shell go list -f {{.Dir}} ./...)

.PHONY: generate
generate: install-mockgen ## Executes all go:generate commands in the source code
	${Q} go generate ./...
	${Q} ${MAKE} format # properly formats all go-source generated files

.PHONY: tools
tools: install-goimports install-mockgen ## Installs set of tools required for local development

install-goimports: go.mod ## Installs a Go source code formatter
	${Q} go install golang.org/x/tools/cmd/goimports

install-mockgen: go.mod ## Installs a Go mock generation tool
	${Q} go install github.com/golang/mock/mockgen
