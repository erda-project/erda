.PHONY: build clean
build:
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