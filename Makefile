.PHONY: dev build test test-cover test-perf bench lint generate ci-test fmt-check fmt-check-go fmt-check-go-staged fmt-check-go-head fmt vuln-check check pre-commit pre-push install-hooks embed-stub gen-test-db

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

# Tracked .go files on disk (aligned with CI checkout; skips node_modules).
fmt-check-go:
	@files=$$(git ls-files '*.go'); \
	[ -n "$$files" ] || exit 0; \
	unformatted=$$(echo "$$files" | xargs gofmt -l); \
	[ -z "$$unformatted" ] || { echo "$$unformatted"; exit 1; }

# Staged .go blob content (what git commit will record).
fmt-check-go-staged:
	@staged=$$(git diff --cached --name-only --diff-filter=ACM -- '*.go'); \
	[ -n "$$staged" ] || exit 0; \
	failed=0; \
	for f in $$staged; do \
		[ -f "$$f" ] || continue; \
		t=$$(mktemp); g=$$(mktemp); \
		git show ":$$f" > "$$t"; \
		gofmt < "$$t" > "$$g"; \
		if ! cmp -s "$$t" "$$g"; then echo "gofmt (staged): $$f"; failed=1; fi; \
		rm -f "$$t" "$$g"; \
	done; \
	[ $$failed -eq 0 ] || { echo "Run: gofmt -w <file> && git add <file>"; exit 1; }

# HEAD .go blob content (what git push will send when the index is clean).
fmt-check-go-head:
	@files=$$(git ls-files '*.go'); \
	[ -n "$$files" ] || exit 0; \
	failed=0; \
	for f in $$files; do \
		t=$$(mktemp); g=$$(mktemp); \
		git show "HEAD:$$f" > "$$t" 2>/dev/null || { rm -f "$$t" "$$g"; continue; }; \
		gofmt < "$$t" > "$$g"; \
		if ! cmp -s "$$t" "$$g"; then echo "$$f"; failed=1; fi; \
		rm -f "$$t" "$$g"; \
	done; \
	[ $$failed -eq 0 ] || { echo "Committed .go files are not gofmt-formatted (run: make fmt && git add -u '*.go')"; exit 1; }

fmt-check: fmt-check-go
	cd frontend && npm run format:check

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

# Lightweight gates before git commit/push (matches CI format + security jobs).
# embed-stub must finish before lint; the rest run in parallel.
HOOK_JOBS ?= 4
pre-commit: embed-stub
	@$(MAKE) -j$(HOOK_JOBS) fmt-check-go-staged fmt-check lint lint-frontend vuln-check

pre-push: embed-stub
	@$(MAKE) -j$(HOOK_JOBS) fmt-check-go-head fmt-check lint lint-frontend vuln-check

install-hooks:
	@chmod +x .githooks/pre-commit .githooks/pre-push
	git config core.hooksPath .githooks
	@echo "Installed git hooks from .githooks (pre-commit -> make pre-commit, pre-push -> make pre-push)"

gen-test-db:
	./data/generate.sh
