# 通用目录
WORK_DIR := $(shell pwd)
BUILD_DIR := $(WORK_DIR)/build
TARGET_DIR := $(WORK_DIR)/target
DIST_DIR := $(TARGET_DIR)/dist
LINT_DIR := $(TARGET_DIR)/lint
TEST_DIR := $(TARGET_DIR)/test

# 通用变量
MOD_NAME := $(shell go list -m)
# 默认不开启verbose模式
VERBOSE ?= 0


#################
##@ Help
#################
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m \033[36m[VERBOSE=1]\033[0m\n"} \
	/^[a-zA-Z_0-9-]+:.*?##/ \
	{ printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ \
	{ printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


#################
##@ Development
#################
.PHONY: generate
generate: ## Run generate code
	@echo "===> step generate..."
	@go generate ./...

.PHONY: fmt
fmt: ## Run go fmt against code.
	@echo "===> step fmt..."
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	@echo "===> step vet..."
	@go vet ./...

.PHONY: lint
lint: ## Run go lint (add VERBOSE=1 for -v)
	@echo "===> step lint..."
	@mkdir -p $(LINT_DIR)
	@$(eval VERBOSE_FLAG = $(if $(filter 1,$(VERBOSE)),-v,))
	@golangci-lint run --output.junit-xml.path=$(LINT_DIR)/lint.xml --output.html.path=$(LINT_DIR)/lint.html \
	$(VERBOSE_FLAG)

.PHONY: test
test: ## Run go test (add VERBOSE=1 for -v)
	@echo "===> step test..."
	@mkdir -p $(TEST_DIR)
	@$(eval VERBOSE_FLAG = $(if $(filter 1,$(VERBOSE)),-v,))
	@go test -gcflags=all=-l $$(go list ./... | grep -v /e2e) -covermode=count -coverprofile=$(TEST_DIR)/cover.out $(VERBOSE_FLAG)
	@$(eval MODULE_DIR = $(shell basename $(WORK_DIR)))
	@gocover diff --cover-profile=$(TEST_DIR)/cover.out -o=$(TEST_DIR) --baseline=0.0 $(VERBOSE_FLAG)

.PHONY: clean
clean: ## Clean up the project.
	@echo "===> step clean..."
	@rm -rf $(TARGET_DIR)


#################
##@ Dist
#################
### 编译信息
VERSION ?= V0.0.1
PLATFORMS ?= linux/amd64 windows/386
COMMIT := $(shell git rev-parse --short HEAD)
LDFLAGS = '-w -X $(MOD_NAME)/internal/cmd.Release=$(VERSION) -X $(MOD_NAME)/internal/cmd.Commit=$(COMMIT)'

.PHONY: build
build: fmt vet generate $(PLATFORMS) ## Build binary
$(PLATFORMS):
	@$(eval DISTTYPE = $(subst /, ,$@))
	@$(eval BUILD_GOOS = $(word 1, $(DISTTYPE)))
	@$(eval BUILD_ARCH = $(word 2, $(DISTTYPE)))
	@echo Building for $(BUILD_GOOS) platform, arch is $(BUILD_ARCH)...
	@mkdir -p $(DIST_DIR)
	@GOOS=$(BUILD_GOOS) GOARCH=$(BUILD_ARCH) GO111MODULE=on CGO_ENABLED=0 go build -trimpath -mod vendor \
-ldflags $(LDFLAGS) -o $(DIST_DIR) ./cmd/...


#################
##@ Dependencies
#################
### 版本信息
GOLANGCI_LINT_VERSION ?= v2.5.0
MOCKGEN_VERSION ?= v0.5.1

.PHONY: depend
depend: golangci-lint mockgen ## Set dependencies

golangci-lint:
	@echo "===> install golangci-lint..."
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

mockgen:
	@echo "===> install mockgen..."
	@go install go.uber.org/mock/mockgen@$(MOCKGEN_VERSION)