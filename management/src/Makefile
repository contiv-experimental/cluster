.PHONY: build checks check-format check-code clean clean-generate deps generate unit-test

work_dir:=/go/src/github.com/contiv/cluster/management/src
docker_run:=docker run --rm -u `id -u`:`id -g` -v `pwd`:$(work_dir) \
	-v $(GOPATH)/bin:/go/bin -w $(work_dir) golang:1.4

deps:
	@echo "checking and downloading dependencies"
	@go get github.com/tools/godep
	@go get github.com/golang/lint/golint
	@go get golang.org/x/tools/cmd/stringer
	@go get github.com/golang/mock/gomock
	@go get github.com/golang/mock/mockgen
	@echo "done checking and downloading dependencies"

checks: deps clean-generate check-format check-code

generate: deps
	@echo "auto generating files"
	@mkdir -p mock
	@go generate ./...
	@echo "done auto generating files"

clean-generate:
	@echo "cleaning auto-generated files"
	@rm -f inventory/*_string.go
	@rm -f clusterm/*_string.go
	@rm -f monitor/*_string.go
	@rm -f inventory/*_mock.go
	@echo "done cleaning auto-generated files"

check-format:
	@echo "checking format..."
	@test -z "$$(gofmt -l . | grep -v Godeps/_workspace/src/ | tee /dev/stderr)"
	@echo "done checking format..."

check-code:
	@echo "checking code..."
	@test -z "$$(golint ./... | tee /dev/stderr)"
	@godep go vet ./...
	@echo "done checking code..."

build: checks generate
	@echo "building..."
	@$(docker_run) godep go install ./...
	@make clean-generate
	@echo "done building..."

unit-test: checks generate
	@echo "running unit-tests"
	@$(docker_run) sh -c "godep go list ./... | xargs -n1 godep go test"
	@make clean-generate
	@echo "done running unit-tests"

clean: deps
	@echo "running clean"
	@godep go clean -i -v ./...
	@echo "done running clean"
