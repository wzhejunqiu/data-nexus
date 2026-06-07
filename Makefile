.PHONY: dev build test test-cover test-coverage test-coverage-go test-coverage-frontend test-coverage-html test-coverage-go-html test-perf bench lint generate sync-version ci-test fmt-check fmt-check-go fmt-check-go-staged fmt-check-go-head fmt vuln-check check pre-commit pre-push install-hooks embed-stub gen-test-db test-db-up test-db-down test-db-integration

COVERAGE_DIR := coverage
GO_COVERAGE := $(COVERAGE_DIR)/go.out
FRONTEND_COVERAGE := $(COVERAGE_DIR)/frontend

UNAME_S := $(shell uname -s)
WAILS_TAGS :=
ifeq ($(UNAME_S),Linux)
WAILS_TAGS := -tags webkit2_41
endif

dev: sync-version
	wails dev $(WAILS_TAGS)

build: sync-version
	wails build $(WAILS_TAGS)

test: sync-version
	go test -short ./...

test-cover: sync-version
	go test -short ./internal/... -cover

# 测试覆盖率（Go + 前端）
test-coverage: test-coverage-go test-coverage-frontend

# Go 测试覆盖率：生成 profile 并在终端输出各包汇总
test-coverage-go: sync-version
	@mkdir -p $(COVERAGE_DIR)
	go test -short ./internal/... -coverprofile=$(GO_COVERAGE) -covermode=atomic
	go tool cover -func=$(GO_COVERAGE)

# 前端测试覆盖率（Vitest，输出 coverage/frontend/）
test-coverage-frontend:
	cd frontend && npm run test:coverage

# 测试覆盖率 HTML 报告（Go + 前端）
test-coverage-html: test-coverage-go-html test-coverage-frontend
	@echo "Go:       $(COVERAGE_DIR)/go.html"
	@echo "Frontend: $(FRONTEND_COVERAGE)/index.html"

test-coverage-go-html: test-coverage-go
	go tool cover -html=$(GO_COVERAGE) -o $(COVERAGE_DIR)/go.html

test-perf:
	go test ./internal/driver/sqlite/... ./internal/service/... -run 'TestPerf' -count=1

bench:
	go test ./internal/driver/sqlite/... -bench=. -benchtime=3x -run='^$$'

generate: embed-stub
	wails generate module

# Single source of truth: wails.json info.productVersion.
# Writes generated internal/version/product.txt (gitignored) and syncs frontend/package.json.
sync-version:
	@ver=$$(node -pe "require('./wails.json').info.productVersion"); \
	printf '%s\n' "$$ver" > internal/version/product.txt; \
	node -e "const fs=require('fs'); const p=require('./frontend/package.json'); p.version='$$ver'; fs.writeFileSync('frontend/package.json', JSON.stringify(p,null,2)+'\n')"; \
	cd frontend && npm install --package-lock-only --silent

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
	@$(MAKE) sync-version

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

# Prefer a running docker compose; else podman compose. Override: make test-db-up COMPOSE="podman compose"
COMPOSE ?= $(shell \
	docker_ready=0; podman_ready=0; \
	if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1 && docker info >/dev/null 2>&1; then \
		docker_ready=1; \
	fi; \
	if command -v podman >/dev/null 2>&1 && podman compose version >/dev/null 2>&1 && podman info >/dev/null 2>&1; then \
		podman_ready=1; \
	fi; \
	if [ "$$docker_ready" = 1 ]; then printf '%s\n' 'docker compose'; \
	elif [ "$$podman_ready" = 1 ]; then printf '%s\n' 'podman compose'; \
	fi)
COMPOSE_FILE := docker-compose.test.yml

define assert-compose-runtime
	@if [ -z "$(COMPOSE)" ]; then \
		echo "No running container runtime found (need docker compose or podman compose)."; \
		if command -v docker >/dev/null 2>&1; then \
			if ! docker info >/dev/null 2>&1; then \
				echo "  Docker: installed but daemon/VM not running — start Docker Desktop."; \
			fi; \
		fi; \
		if command -v podman >/dev/null 2>&1; then \
			if ! podman info >/dev/null 2>&1; then \
				echo "  Podman: installed but machine not running — try: podman machine start"; \
			fi; \
		fi; \
		exit 1; \
	fi
endef

test-db-up:
	$(assert-compose-runtime)
	@echo "Using $(COMPOSE)"
	$(COMPOSE) -f $(COMPOSE_FILE) up -d --wait
	@chmod +x data/testdb/seed.sh
	@./data/testdb/seed.sh $(COMPOSE_FILE) $(COMPOSE)

test-db-down:
	$(assert-compose-runtime)
	$(COMPOSE) -f $(COMPOSE_FILE) down -v

test-db-integration: test-db-up
	TEST_POSTGRES_DSN='postgres://test:test@127.0.0.1:5432/testdb?sslmode=disable' \
	TEST_MYSQL_DSN='test:test@tcp(127.0.0.1:3306)/testdb?parseTime=true' \
	go test ./internal/driver/postgres/... ./internal/driver/mysql/... -run Integration -count=1
	$(MAKE) test-db-down
