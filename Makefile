.PHONY: check fmt-check vet tidy verify test race

check: fmt-check vet tidy verify test race

fmt-check:
	test -z "$$(gofmt -l .)"

vet:
	go vet ./...

tidy:
	go mod tidy -diff

verify:
	go mod verify

test:
	go test ./...

race:
	go test -race ./...
