.DEFAULT_GOAL := build

.PHONY: build
build:
	go build -o ./slash-prompt -ldflags="-X 'main.Version=development'" .

.PHONY: build-docker
build-docker:
	docker build -t slash-prompt:development .

.PHONY: test
test:
	rm -rf .coverdata/unit
	mkdir -p .coverdata/unit
	go test -cover ./internal/... -args -test.gocoverdir="`pwd`/.coverdata/unit"

.PHONY: e2e
e2e:
	./scripts/setup-e2e.sh
	rm -rf ./e2e/_output/*
	go build -o ./slash-prompt-e2e -ldflags="-X 'main.Version=development'" .
	go test ./e2e/...

.PHONY: e2e-update
e2e-update:
	./scripts/setup-e2e.sh
	go build -o ./slash-prompt-e2e .
	go test ./e2e/... -update

.PHONY: clean
clean:
	rm -f ./slash-prompt ./slash-prompt-e2e
	rm -rf .coverdata
	rm -rf ./e2e/_output/*