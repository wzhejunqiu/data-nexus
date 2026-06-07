.PHONY: dev build test lint generate ci-test fmt-check fmt vuln-check check

dev:
	wails dev

build:
	wails build

test:
	go test ./...

generate:
	wails generate module

fmt-check:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

fmt:
	gofmt -w .

ci-test: generate
	cd frontend && npm ci && npm run format:check && npm test && npm run build
	go test ./...

lint:
	golangci-lint run ./...

vuln-check:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	cd frontend && npm audit --omit=dev --audit-level=high

check: fmt-check lint vuln-check ci-test
