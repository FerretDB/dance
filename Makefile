all: fmt test

help:                                  ## Display this help message
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
		awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'

env-up: env-up-detach                  ## Start development environment
	docker-compose logs --follow

env-up-detach:
	docker-compose up --always-recreate-deps --force-recreate --remove-orphans --renew-anon-volumes --detach

env-down:                              ## Stop development environment
	docker-compose down --remove-orphans

init:                                  ## Install development tools
	go mod tidy
	cd tools && go mod tidy && go generate -tags=tools -x

fmt: bin/gofumpt                       ## Format code
    # skip submodules
	bin/gofumpt -w ./cmd/ ./internal/

lint: bin/golangci-lint                ## Run linters
	bin/golangci-lint run --config=.golangci-required.yml
	bin/golangci-lint run --config=.golangci.yml

bin/golangci-lint:
	$(MAKE) init
