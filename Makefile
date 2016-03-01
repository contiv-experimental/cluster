demo: stop start
stop:
	CONTIV_NODES=$${CONTIV_NODES:-3} vagrant destroy -f
start:
	CONTIV_NODES=$${CONTIV_NODES:-3} CONTIV_SRV_INIT=1 vagrant up
svc-provision:
	CONTIV_NODES=$${CONTIV_NODES:-3} CONTIV_SRV_INIT=1 \
				 vagrant provision --provision-with ansible
svc-cleanup:
	CONTIV_NODES=$${CONTIV_NODES:-3} CONTIV_ANSIBLE_PLAYBOOK="./vendor/ansible/cleanup.yml" \
				 vagrant provision --provision-with ansible
