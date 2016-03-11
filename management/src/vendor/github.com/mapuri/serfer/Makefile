.PHONY: checks check-format check-code deps unit-test test-app

all: checks test-app

deps:
	go get github.com/tools/godep
	go get github.com/golang/lint/golint

checks: deps check-format check-code unit-test
	
check-format:
	@echo "checking format..."
	test -z "$$(gofmt -l . | grep -v Godeps/_workspace/src/ | tee /dev/stderr)"
	@echo "done checking format..."

check-code:
	@echo "checking code..."
	test -z "$$(golint ./... | tee /dev/stderr)"
	go vet ./...
	@echo "done checking code..."

unit-test:
	#godep go test ./...
	go list ./... | xargs -n1 godep go test

test-app:
	godep go install ./test-app

clean: deps
	godep go clean -i -v ./...
