.PHONY: build build-docker-img build-docker-shell checks check-format check-code clean-generate clean-tar deps generate host-system-test unit-test system-test tar release

work_dir:=/go/src/github.com/contiv/cluster/management/src
docker_buildargs:= \
	--build-arg "uid=$$(id -u)" \
	--build-arg "user=$$(id -un)" \
	--build-arg "work_dir=$(work_dir)" \
	$(if $(HTTP_PROXY), --build-arg "HTTP_PROXY=$(HTTP_PROXY)") \
	$(if $(HTTPS_PROXY), --build-arg "HTTPS_PROXY=$(HTTPS_PROXY)") \
	$(if $(http_proxy), --build-arg "http_proxy=$(http_proxy)") \
	$(if $(https_proxy), --build-arg "https_proxy=$(https_proxy)")
docker_img:=cluster-dev$(if $(GIT_BRANCH),:$(GIT_BRANCH))
docker_run:=docker run --rm -u `id -un`:`id -un` -v `pwd`:$(work_dir) \
	-w $(work_dir) "$(docker_img)"
docker_run_interactive:=docker run -it --rm -u `id -un`:`id -un` -v `pwd`:$(work_dir) \
	-w $(work_dir) "$(docker_img)"
unittest_tag:=unittest
systemtest_tag:=systemtest
all_tags:=$(unittest_tag) $(systemtest_tag)
all_packages:=go list -tags '$(all_tags)' ./... | grep -v vendor/
# filtering systemtest packages as below helps filter out unittest packages
# that need code generation, which we don't need for running systemtests
systemtest_packages:=go list -tags '$(systemtest_tag)'  -f '{{.ImportPath}}:{{.TestGoFiles}}' ./... | \
		grep -v vendor/ | \
		grep -v '\[\]' | \
		awk -F: '{print $$1}'

# set CONTIV_TEST_FLAGS in environment to run specific system tests
# Eg. CONTIV_TEST_FLAGS="-check.f=TestDecommission*"
CONTIV_TEST_FLAGS:= \
	$(if $(CONTIV_TEST_FLAGS),"$(CONTIV_TEST_FLAGS)", "")

deps:
	@echo "checking and downloading dependencies"
	@go get github.com/tools/godep
	@go get github.com/golang/lint/golint
	@go get golang.org/x/tools/cmd/stringer
	@go get github.com/golang/mock/gomock
	@go get github.com/golang/mock/mockgen
	@echo "done checking and downloading dependencies"

checks: deps clean-generate check-format check-code

godeps-save:
	$(all_packages) | xargs godep save

generate: deps
	@echo "auto generating files"
	@mkdir -p mock
	@$(all_packages) | xargs -n1 go generate
	@echo "done auto generating files"

clean-generate:
	@echo "cleaning auto-generated files"
	@rm -f inventory/*_string.go
	@rm -f clusterm/*_string.go
	@rm -f monitor/*_string.go
	@rm -f mock/*_mock.go
	@echo "done cleaning auto-generated files"

check-format:
	@echo "checking format..."
	@test -z "$$(gofmt -l . | grep -v vendor/ | tee /dev/stderr)"
	@echo "done checking format..."

check-code:
	@echo "checking code..."
	@test -z "$$($(all_packages) | xargs -n1 golint | tee /dev/stderr)"
	@$(all_packages) | xargs -n1 go vet
	@echo "done checking code..."

build-docker-img:
	docker build -t "$(docker_img)" $(docker_buildargs) .

build-docker-shell: build-docker-img
	@$(docker_run_interactive) bash

build: build-docker-img
	@echo "building test image..."
	@$(docker_run) bash -c "make checks generate && \
		($(all_packages) | xargs -n1 go install) && \
		make clean-generate"
	@echo "done building test image..."

unit-test: build-docker-img
	@echo "running unit-tests"
	@$(docker_run) bash -c "make checks generate && \
		($(all_packages) | \
		xargs -n1 -I'{}' go test -tags '$(unittest_tag)' '{}' -check.v $(CONTIV_TEST_FLAGS)) && \
		make clean-generate"
	@echo "done running unit-tests"

host-system-test: checks
	@$(systemtest_packages) | \
		xargs -n1 -I'{}' go test -tags '$(systemtest_tag)' '{}' -check.v $(CONTIV_TEST_FLAGS)

system-test:
	@echo "running system-tests"
	CONTIV_NODES=2 time vagrant up; \
	time vagrant ssh cluster-node1 -c \
	"export CONTIV_SOE=${CONTIV_SOE}; \
	 export GOPATH=/opt/gopath/; \
	 export PATH=\$$PATH:\$$GOPATH/bin; \
	 export CONTIV_NODE_IPS='192.168.2.10,192.168.2.11'; \
	 if [ \"${http_proxy}\" != \"\" ]; then export http_proxy=${http_proxy}; fi; \
	 if [ \"${https_proxy}\" != \"\" ]; then export https_proxy=${https_proxy}; fi; \
	 cd \$$GOPATH/src/github.com/contiv/cluster/management/src; \
	 CONTIV_TEST_FLAGS=$(CONTIV_TEST_FLAGS) make host-system-test"; \
	if [ "$$?" != "0" ]; then res=1; else res=0; fi; \
	if [ "$${res}" = "0" -o "${CONTIV_SOE}" = "" ]; then CONTIV_NODES=2 vagrant destroy -f; fi; \
	exit $${res}
	@echo "done running system-tests"

# We are using date based versioning, so for consistent version during a build
# we evaluate and set the value of version once in a file and use it in 'tar'
# and 'release' targets.
NAME:=cluster
VERSION_FILE:=/tmp/$(NAME)-version
VERSION:=`cat $(VERSION_FILE)`
TAR_EXT:=tar.bz2
TAR_FILENAME:=$(NAME)-$(VERSION).$(TAR_EXT)
TAR_LOC:=.
TAR_FILE:=$(TAR_LOC)/$(TAR_FILENAME)

tar: build
	@echo "building tarball"
	@echo "v0.0.0-`date -u +%m-%d-%Y.%H-%M-%S.UTC`" > $(VERSION_FILE)
	@tar -jcf $(TAR_FILE) -C ./bin clusterctl clusterm
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
