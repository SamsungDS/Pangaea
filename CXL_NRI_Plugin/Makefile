#   Copyright The containerd Authors.

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

GO_CMD     := go
GO_BUILD   := $(GO_CMD) build
GO_INSTALL := $(GO_CMD) install
GO_TEST    := $(GO_CMD) test
GO_LINT    := golint -set_exit_status
GO_FMT     := gofmt
GO_VET     := $(GO_CMD) vet

GO_BUILD_FLAGS ?=

GO_MODULES := $(shell $(GO_CMD) list ./...)

GOLANG_CILINT := golangci-lint

RESOLVED_PWD  := $(shell realpath $(shell pwd))
BUILD_PATH    := $(RESOLVED_PWD)/build
BIN_PATH      := $(BUILD_PATH)/bin

PLUGINS := \
	$(BIN_PATH)/cxl

ifneq ($(V),1)
  Q := @
endif

#
# top-level targets
#

all: build build-plugins

build: build-check

clean: clean-plugins

allclean: clean clean-cache

FORCE:

#
# build targets
#

build-plugins: $(PLUGINS)

build-check:
	$(Q)$(GO_BUILD) -v $(GO_MODULES)

mod-tidy:
	$(Q)$(GO_CMD) mod tidy
	$(Q)./scripts/go-mod-tidy

#
# clean targets
#

clean-plugins:
	$(Q)rm -f $(PLUGINS)

clean-cache:
	$(Q)$(GO_CMD) clean -cache -testcache

#
# plugins build targets
#

$(BIN_PATH)/% build/bin/%: FORCE
	$(Q)echo "Building $@..."; \
	$(GO_BUILD) -C cmd/$* -o $(abspath $@) $(GO_BUILD_FLAGS) .

#
# other validation targets
#

fmt format:
	$(Q)$(GO_FMT) -s -d -e .

lint:
	$(Q)$(GO_LINT) -set_exit_status ./...

vet:
	$(Q)$(GO_VET) ./...

golangci-lint:
	$(Q)$(GOLANG_CILINT) run

