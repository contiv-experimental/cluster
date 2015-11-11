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

# We are using date based versioning, so for consistent version during a build
# we evaluate and set the value of version once in a file and use it in 'tar'
# and 'release' targets.
NAME := cluster
VERSION_FILE := /tmp/$(NAME)-version
VERSION := `cat $(VERSION_FILE)`
TAR_EXT := tar.bz2
TAR_FILENAME := $(NAME)-$(VERSION).$(TAR_EXT)
TAR_LOC := .
TAR_FILE := $(TAR_LOC)/$(TAR_FILENAME)

tar: build
	@echo "building tarball"
	@echo "v0.0.0-`date -u +%m-%d-%Y.%H-%M-%S.UTC`" > $(VERSION_FILE)
	@tar -jcf $(TAR_FILE) -C $(GOPATH)/bin clusterctl clusterm
	@echo "done building tarball"

clean-tar:
	@echo "cleaning tarball"
	@rm -f $(TAR_LOC)/*.$(TAR_EXT)
	@echo "done cleaning tarball"

# GITHUB_USER and GITHUB_TOKEN are needed be set to run github-release
release: tar
	@echo "publishing release"
	@go get github.com/aktau/github-release
	@latest_tag=$$(git describe --tags `git rev-list --tags --max-count=1`); \
		comparison="$$latest_tag..HEAD"; \
		changelog=$$(git log $$comparison --oneline --no-merges --reverse); \
		if [ -z "$$changelog" ]; then echo "No new changes to release!"; exit 0; fi; \
		set -x; \
		( ( github-release -v release -p -r cluster -t $(VERSION) -d "**Changelog**<br/>$$changelog" ) && \
		( github-release -v upload -r cluster -t $(VERSION) -n $(TAR_FILENAME) -f $(TAR_FILE) || \
		github-release -v delete -r cluster -t $(VERSION) ) ) || exit 1
	@make clean-tar
	@echo "done publishing release"
