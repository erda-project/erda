.PHONY: build clean
build: auto-update-infra-for-master
	@./build.sh build
	@go mod tidy
	@echo "" && make format
clean:
	@./build.sh clean

format:
	@echo "run gofmt && goimports"
	@GOFILES=$$(find . -name "*.go"); \
	for path in $${GOFILES}; do \
		gofmt -w -l $${path}; \
		goimports -w -l $${path}; \
	done;

auto-update-infra-for-master:
	@if [[ "$$(git rev-parse --abbrev-ref HEAD)" == "master" ]]; then \
		make update-infra; \
	fi;

update-infra:
	echo "update infra and gohub"
	go get -u github.com/erda-project/erda-infra
	go get -u github.com/erda-project/erda-infra/tools/gohub