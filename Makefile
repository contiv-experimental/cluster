demo: stop start

stop:
	CONTIV_NODES=$${CONTIV_NODES:-3} CONTIV_SRV_INIT=1 vagrant destroy -f

start:
	CONTIV_NODES=$${CONTIV_NODES:-3} CONTIV_SRV_INIT=1 vagrant up

vendor-ansible:
	git subtree pull --prefix vendor/ansible https://github.com/contiv/ansible HEAD --squash
