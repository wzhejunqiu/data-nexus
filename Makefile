.PHONY: dev build test test-cover test-perf bench lint generate ci-test fmt-check fmt vuln-check check pre-push install-hooks embed-stub gen-test-db

UNAME_S := $(shell uname -s)
WAILS_TAGS :=
ifeq ($(UNAME_S),Linux)
WAILS_TAGS := -tags webkit2_41
endif

dev:
	wails dev $(WAILS_TAGS)

build:
	wails build $(WAILS_TAGS)

test:
	go test -short ./...

test-cover:
	go test -short ./internal/... -cover

test-perf:
	go test ./internal/driver/sqlite/... ./internal/service/... -run 'TestPerf' -count=1

bench:
	go test ./internal/driver/sqlite/... -bench=. -benchtime=3x -run='^$$'

generate: embed-stub
	wails generate module

fmt-check:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

fmt:
	gofmt -w .

embed-stub:
	@mkdir -p frontend/dist
	@test -f frontend/dist/index.html || printf '%s\n' '<!doctype html><html><head></head><body></body></html>' > frontend/dist/index.html

lint-frontend:
	cd frontend && npm run lint

ci-test: generate
	cd frontend && npm ci && npm run format:check && npm run lint && npm test && npm run build
	go test -short ./...

lint: embed-stub
	golangci-lint run ./internal/... .

vuln-check:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	cd frontend && npm audit --omit=dev --audit-level=high

check: fmt-check lint vuln-check ci-test

# Lightweight gate before git push (matches CI format + security jobs).
pre-push: embed-stub fmt-check lint vuln-check

install-hooks:
	@chmod +x .githooks/pre-push
	git config core.hooksPath .githooks
	@echo "Installed git hooks from .githooks (pre-push -> make pre-push)"

gen-test-db:
	./data/generate.sh
