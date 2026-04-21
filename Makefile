PROJECT ?=
SLUG ?=
ADDR ?= :6345
WORKSPACE ?=

.PHONY: init check-docs check-repo test build-web build run release-package new-history new-plan

init:
	@if [ -z "$(PROJECT)" ]; then echo "用法: make init PROJECT=项目名"; exit 1; fi
	./scripts/init-project.sh "$(PROJECT)"

check-docs:
	./scripts/check-docs.sh

check-repo:
	./scripts/check-docs.sh
	./scripts/check-repo-hygiene.sh

ci:
	./scripts/ci.sh

test:
	go test ./...
	npm --prefix apps/web run build

build-web:
	npm --prefix apps/web run build

build: build-web
	go build -o bin/daymine ./apps/daymine/cmd/daymine

run: build-web
	@if [ -z "$(WORKSPACE)" ]; then \
		go run ./apps/daymine/cmd/daymine --addr "$(ADDR)"; \
	else \
		go run ./apps/daymine/cmd/daymine --addr "$(ADDR)" --workspace "$(WORKSPACE)"; \
	fi

release-package:
	./scripts/release-package.sh

new-history:
	@if [ -z "$(SLUG)" ]; then echo "用法: make new-history SLUG=变更名"; exit 1; fi
	./scripts/new-history.sh "$(SLUG)"

new-plan:
	@if [ -z "$(SLUG)" ]; then echo "用法: make new-plan SLUG=计划名"; exit 1; fi
	./scripts/new-exec-plan.sh "$(SLUG)"
