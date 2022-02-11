PACKAGE=github.com/epam/edp-common/pkg/config
CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/dist
BIN_NAME=go-binary

HOST_OS:=$(shell go env GOOS)
HOST_ARCH:=$(shell go env GOARCH)

VERSION=$(shell cat ${CURRENT_DIR}/VERSION)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_TAG=$(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
KUBECTL_VERSION=$(shell go list -m all | grep k8s.io/client-go| cut -d' ' -f2)

override LDFLAGS += \
  -X ${PACKAGE}.version=${VERSION} \
  -X ${PACKAGE}.buildDate=${BUILD_DATE} \
  -X ${PACKAGE}.gitCommit=${GIT_COMMIT} \
  -X ${PACKAGE}.kubectlVersion=${KUBECTL_VERSION}\

ifneq (${GIT_TAG},)
LDFLAGS += -X ${PACKAGE}.gitTag=${GIT_TAG}
endif

.DEFAULT_GOAL:=help
# set default shell
SHELL=/bin/bash -o pipefail -o errexit
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

# Run tests
test: fmt vet
	KUBECONFIG=${CURRENT_DIR}/hack/kubecfg-stub.yaml go test ./... -coverprofile=coverage.out `go list ./...`

fmt:  ## Run go fmt
	go fmt ./...

vet:  ## Run go vet
	go vet ./...

lint: ## Run go lint
	golangci-lint run

.PHONY: build
build: clean ## build operator's binary
	CGO_ENABLED=0 GOOS=${HOST_OS} GOARCH=${HOST_ARCH} go build -v -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${BIN_NAME} ./cmd/manager/main.go

.PHONY: clean
clean:  ## clean up
	-rm -rf ${DIST_DIR}

# use https://github.com/git-chglog/git-chglog/
.PHONY: changelog
changelog: ## generate changelog
ifneq (${NEXT_RELEASE_TAG},)
	@git-chglog --next-tag v${NEXT_RELEASE_TAG} -o CHANGELOG.md v2.7.0..
else
	@git-chglog -o CHANGELOG.md v2.7.0..
endif

.PHONY: api-docs
api-docs: ## generate CRD docs
	crdoc --resources deploy-templates/crds --output docs/api.md

.PHONY: gen-mocks
gen-mocks: gen-platform-service-mock gen-sonar-client-mock gen-sonar-service-mock gen-k8s-clients-mock gen-openshift-clients-mock

.PHONY: gen-platform-service-mock
gen-platform-service-mock:
	docker run -v `pwd`:/src -w /src vektra/mockery:v2.9 --case snake --name Service --dir ./pkg/service/platform --output mocks/platform --outpkg mock --exported --filename mock_service.go

.PHONY: gen-sonar-service-mock
gen-sonar-service-mock:
	docker run -v `pwd`:/src -w /src vektra/mockery:v2.9 --case snake --name ServiceInterface --dir ./pkg/service/sonar --output mocks/service --outpkg mock --exported --filename sonar_service.go

.PHONY: gen-sonar-client-mock
gen-sonar-client-mock:
	docker run -v `pwd`:/src -w /src vektra/mockery:v2.9 --case snake --name ClientInterface --dir ./pkg/service/sonar --output mocks/client --outpkg mock --exported --filename sonar_client.go

.PHONY: gen-k8s-clients-mock
gen-k8s-clients-mock:
	docker run -v `pwd`:/src -w /src vektra/mockery:v2.9 --case snake --all --dir ./pkg/service/platform/kubernetes --output mocks/k8s --outpkg mock --exported

.PHONY: gen-openshift-clients-mock
gen-openshift-clients-mock:
	docker run -v `pwd`:/src -w /src vektra/mockery:v2.9 --case snake --all --dir ./pkg/service/platform/openshift --output mocks/openshift --outpkg mock --exported
