.PHONY: build gen test lint fmt vet tidy e2e install

# Build the binary with version metadata.
VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o gr4vy .

# Regenerate the command surface from the committed spec + gr4vy-go types.
gen:
	go generate ./...

# Refresh the committed spec from the hosted source, then regenerate.
gen-refresh:
	curl -fsSL https://gr4vy.github.io/openapi/core/openapi.json -o internal/spec/openapi.json
	go generate ./...

test:
	go test ./...

# Live e2e suite (requires PRIVATE_KEY or a private_key.pem at the repo root).
e2e:
	go test ./test/... -count=1 -v

vet:
	go vet ./...

fmt:
	gofmt -w .

lint: vet
	gofmt -l .

tidy:
	go mod tidy

install:
	go install -ldflags "$(LDFLAGS)" .
