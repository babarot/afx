ifdef DEBUG
	GOFLAGS := -gcflags="-N -l"
else
	GOFLAGS :=
endif

GO        ?= go
TAGS      :=
LDFLAGS   :=

GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null || echo "canary")
GIT_DIRTY  = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

LDFLAGS += -X github.com/babarot/afx/cmd.BuildSHA=${GIT_SHA}
LDFLAGS += -X github.com/babarot/afx/cmd.GitTreeState=${GIT_DIRTY}

ifneq ($(GIT_TAG),)
	LDFLAGS += -X github.com/babarot/afx/cmd.BuildTag=${GIT_TAG}
endif

all: build

.PHONY: build
build:
	$(GO) install $(GOFLAGS) -ldflags '$(LDFLAGS)'

.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: docs
docs:
	@python3 -m pip install --upgrade pip
	@python3 -m pip install -r ./docs/requirements.txt
	@python3 -m mkdocs serve
