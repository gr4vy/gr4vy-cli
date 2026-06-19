.PHONY: build gen test lint fmt vet tidy e2e install

# Build the binary with version metadata.
VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o gr4vy .

# Regenerate the command surface from the gr4vy-go SDK types.
gen:
	go generate ./...

# Bump gr4vy-go to its latest release, then regenerate (mirrors the regen CI job).
gen-refresh:
	go get -u github.com/gr4vy/gr4vy-go@latest
	go mod tidy
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
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed on:"; echo "$$unformatted"; exit 1; \
	fi

tidy:
	go mod tidy

install:
	go install -ldflags "$(LDFLAGS)" .
