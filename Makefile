ROOT := $(CURDIR)
export GOCACHE := $(ROOT)/.cache/go-build
export GOMODCACHE := $(ROOT)/.cache/go-mod

.PHONY: frontend check test build prototype
frontend:
	npm run build:frontend
check: frontend
	npm run check
	cd apps/desktop && go vet ./...
test: frontend
	cd apps/desktop && go test -race ./...
build: frontend
	cd apps/desktop && go build -tags desktop,production -o build/bin/pdf-tools .
prototype: frontend
	cd apps/desktop && bash scripts/package-macos.sh
