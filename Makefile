.PHONY: check fmt-check vet lint test race tidy verify vuln

check: fmt-check vet lint test race tidy verify

fmt-check:
	test -z "$(shell gofmt -l .)"

vet:
	go vet ./...

# TODO: Replace this compatibility alias with a pinned external linter when one is approved.
lint: vet

test:
	go test ./...

race:
	go test -race ./...

tidy:
	go mod tidy -diff

verify:
	go mod verify

# TODO: Run a pinned external vulnerability scanner when external development tools are approved.
vuln:
	@true
