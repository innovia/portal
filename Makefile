# Set default shell to bash
SHELL := /bin/bash -o pipefail -o errexit -o nounset

GO_RELEASER_VERSION=v1.2.5
GO_BUILD_FLAGS=--ldflags '-extldflags "-static"'

NAME=portal
BINARY=${NAME}


# Project variables
DOCKER_REGISTRY ?= innovia
DOCKER_IMAGE = ${DOCKER_REGISTRY}/${NAME}

# Build variables
VERSION ?= $(shell echo `git describe --tags --exact-match` | tr '[/]' '-')
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE ?= $(shell date +%FT%T%z)
LDFLAGS += -X main.version=${VERSION} -X main.commitHash=${COMMIT_HASH} -X main.buildDate=${BUILD_DATE}
export CGO_ENABLED ?= 1
export GOOS = $(shell go env GOOS)
ifeq (${VERBOSE}, 1)
	GOARGS += -v
endif

# Docker variables
DOCKER_TAG ?= ${VERSION}

GOLANG_VERSION = 1.19


# Colors for the printf
RESET = $(shell tput sgr0)
COLOR_WHITE = $(shell tput bold setaf 7)
COLOR_BLUE = $(shell tput setaf 4)
COLOR_YELLOW = $(shell tput setaf 3)


#-----------------------------------------------------------------------------------------------------------------------
# Rules
#-----------------------------------------------------------------------------------------------------------------------
default: build

build:
	CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -o bin/portal

release/tag: VERSION?=v$(shell cat VERSION)
release/tag:
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

release/dry-run:
	go run github.com/goreleaser/goreleaser@$(GO_RELEASER_VERSION) release --snapshot --rm-dist

# goreleaser will build the docker images and push
release:
	go run github.com/goreleaser/goreleaser@$(GO_RELEASER_VERSION) release --rm-dist

test:
	gotestsum --format testname ./...

certs-cleanup:
	${call print_warning, "Will delete all certificates and keys from certs folder!"}
	@echo "Are you sure? [y/N] " && read ans && if [ $${ans:-'N'} = 'y' ]; then printf "Deleting certs folder..." && rm -rf certs ;echo "done!"; fi

certs-gen-server:
	${call print, "Generating Server and CA Certificates"}
	@./generate_server_certs.sh

certs-gen-client:
	${call print, "Generating Client Certificate"}
	@./generate_client_certs.sh

start-kind:
	${call print, "Starting Kind cluster"}
	kind create cluster --name portal-local-cluster

run-server-in-kind:
	${call print, "Running Portal in Docker with Kind"}
	docker run --net=host -v ~/.kube:/.kube -v $(PWD)/certs:/certs $(DOCKER_IMAGE):$(VERSION) --tls_cert_file /certs/server.crt --tls_private_key_file /certs/server.key --ca_cert_file /certs/ca_cert.pem  --kubeconfig /.kube/config

clean:
	rm -rf ./bin
	rm -rf ./certs
	rm -rf ./dist

help:
	@echo  'Make targets:'
	@echo  '  build              - build the server'
	@echo  '  release/tag        - use VERSION=v0.0.0 release/tag to create a new tag and push - this will trigger a GoReleaser build binaries and Docker push'
	@echo  '  release/dry-run    - run GoReleaser locally and skip pushing Docker images'
	@echo  '  release            - run GoReleaser locally and push to Docker images'
	@echo  '  test               - run unit tests'
	@echo  '  certs-cleanup      - delete certs folder'
	@echo  '  certs-gen-server   - generate server certificates and root CA'
	@echo  '  certs-gen-client   - generate client certificates'
	@echo  '  start-kind         - Run Kubernetes In Docker (Kind)'
	@echo  '  run-server-in-kind - run Portal server in Docker and network into kind'
	@echo  '  clean              - delete bin dist and certs folders'
	@echo  ''
#-----------------------------------------------------------------------------------------------------------------------
# Helpers
#-----------------------------------------------------------------------------------------------------------------------
define print
	@printf "${COLOR_BLUE}==> ${RESET} ${COLOR_WHITE} %-20s ${RESET}\n" $(1)
endef

define print_warning
	@printf "⚠️  ${COLOR_YELLOW}Warning!${RESET} ${COLOR_YELLOW}==> ${RESET} ${COLOR_WHITE} %-20s ${RESET}\n" $(1)
endef

