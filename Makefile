.PHONY: demo demo-stock help start start-stock stop svc-provision svc-cleanup

.DEFAULT_GOAL := help

help: ## This help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

demo: stop start ## Bring up a demo setup with all services started. This tearsdown any existing setup first.

demo-cluster: stop start-cluster ## Bring up a demo setup with just clusterm running.  This is useful for trying out clusterm workflows. This tearsdown any existing setup first.

demo-stock: stop start-stock ## Bring up a demo setup with stock OS box. All packages are installed fresh. This is useful for testing production like environment. This tearsdown any existing setup first.

stop: ## Teardown a demo setup.
	CONTIV_NODES=$${CONTIV_NODES:-3} vagrant destroy -f

start: ## Bring up a demo setup with all services started.
	CONTIV_NODES=$${CONTIV_NODES:-3} CONTIV_REL_CLUSTERM=1 CONTIV_SRV_INIT=1 vagrant up

start-stock: ## Bring up a demo setup with stock OS box. All packages are installed fresh. Useful for testing production like environment.
	CONTIV_BOX="puppetlabs/centos-7.2-64-nocm" CONTIV_BOX_VERSION="1.0.1" make start

svc-provision: ## Rerun ansible provisioning on the exisitng demo setup.
	CONTIV_NODES=$${CONTIV_NODES:-3} CONTIV_REL_CLUSTERM=1 CONTIV_SRV_INIT=1 \
				 vagrant provision --provision-with ansible

svc-cleanup: ## Run cleanup ansible on the existing demo setup.
	CONTIV_NODES=$${CONTIV_NODES:-3} CONTIV_ANSIBLE_PLAYBOOK="./vendor/ansible/cleanup.yml" \
				 vagrant provision --provision-with ansible

start-cluster: ## Bring up a demo setup with just clusterm running on first node. This is useful for trying out clusterm workflows.
	CONTIV_NODES=$${CONTIV_NODES:-3} CONTIV_REL_CLUSTERM=1 vagrant up
