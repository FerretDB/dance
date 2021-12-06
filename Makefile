all: fmt test

help:                                  ## Display this help message
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
		awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'

env-up: env-up-detach                  ## Start development environment
	docker-compose logs --follow

env-up-detach:
	docker-compose up --always-recreate-deps --force-recreate --remove-orphans --renew-anon-volumes --detach
