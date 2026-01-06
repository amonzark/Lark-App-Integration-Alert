GOIMPORTS   := golang.org/x/tools/cmd/goimports@latest
GOVULNCHECK := golang.org/x/vuln/cmd/govulncheck@latest

.PHONY: test dev
test:
	@go test -v -covermode=count -shuffle=on ./...
dev:
	@docker-compose -f docker/docker-compose.yaml up --build --force-recreate --no-deps