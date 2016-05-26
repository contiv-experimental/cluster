# Troubleshooting Guide

This guide provides an overview of common steps to narrow down failures seen during clusterm workflows.

### Clusterm: Under The Hoods

This section details what goes on behind the scenes when a cluster manager workflow is triggered. The section is divided into following sub-sections:
- [Discovery](#discovery): lists details of discovery process
- [Commission](#commission): lists details of the node(s) commission process
- [Decommission](#decommission): lists details of the node(s) decommission process

Clusterm uses [ansible](https://www.ansible.com/get-started) playbooks for handling different workflows. So being able to find the ansible task that failed is crucial to troubleshooting configuration failure. [Parsing Ansible Logs](#parsing-ansible-logs) section walks through a few samples of ansible logs and patterns to look for failures.

For interested reader, the ansible playbooks used by cluster manager are located [here](https://github.com/contiv/ansible). The understanding of ansible playbooks is not necessary for troubleshooting configuration failures.

#### Discovery
The [Discovery workflow](./baremetal.md/#3-provision-additional-nodes-for-discovery) runs [site.yml playbook](https://github.com/contiv/ansible/blob/master/site.yml) with the nodes that are being discovered added to the `cluster-node` host-group.

This playbook provisions [serf](https://www.serfdom.io/) on the nodes.

##### On Success:
Once the configuration job is done, i.e. when `clusterctl job get active` command returns `info for "active" job doesn't exist` error, the node shall become visible in the output of `clusterctl nodes get` command.

##### On Failure:
Once the configuration job is done, i.e. when `clusterctl job get active` returns `info for "active" job doesn't exist` error, the node will still not be visible in the output of `clusterctl nodes get` command.

Refer [Useful Commands](#useful-commands) section to see commands to run and making sense of the logs.

#### Commission
The [Commission workflow](./README.md#commission-a-node) runs [site.yml playbook](https://github.com/contiv/ansible/blob/master/site.yml) with the nodes that are being commissioned added to either `service-master` or `service-worker` host-group as specified in the commission command.

While commission is in progress the node's status is changed to 'Provisioning'.

##### On Success:
Once the configuration job is done, i.e. when `clusterctl job get active` returns `info for "active" job doesn't exist` error, the node status is be updated to `Allocated` in the  `clusterctl nodes get` command.

##### On Failure:
Once the configuration job is done, i.e. when `clusterctl job get active` returns `info for "active" job doesn't exist` error, the node status is be updated to `Unallocated` in the  `clusterctl nodes get` command. The [cleanup.yml playbook](https://github.com/contiv/ansible/blob/master/cleanup.yml) is run on the node to remove any partial configuration remanants.

Refer [Useful Commands](#useful-commands) section to see commands to run and making sense of the logs.

#### Decommission
The [Decommission workflow](./README.md#decommission-a-node) runs [cleanup.yml playbook](https://github.com/contiv/ansible/blob/master/cleanup.yml) on the nodes that are being decommissioned. This shall stop the infra services and cleanup any service state from the node.

While decommission is in progress the node's status is changed to 'Cancelled'.

##### On Success and Failure:
Once the configuration job is done, i.e. `clusterctl job get active` returns `info for "active" job doesn't exist` error, the node status is be updated to `Decommissioned` in the  `clusterctl nodes get` command.

The decommission playbook runs in it's entirety to do the best effort cleanup and is not expected to return failure.

### Useful commands
This section lists the a few useful commands, with sample outputs and common examples of parsing the logs.

#### Clusterm State and Logs 

##### clusterctl nodes get
This command dumps the info for all the nodes in the inventory. The current inventory status of node is shown after `status` field under `Inventory State`. Following is a sample output for a 3 node cluster:

```
[vagrant@cluster-node1 ~]$ clusterctl nodes get
cluster-node1-0: Inventory State
cluster-node1-0:    name: cluster-node1-0
cluster-node1-0:    prev_state: Discovered
cluster-node1-0:    prev_status: Provisioning
cluster-node1-0:    state: Discovered
cluster-node1-0:    status: Allocated
cluster-node1-0: Monitoring State
cluster-node1-0:    label: cluster-node1
cluster-node1-0:    management_address: 192.168.2.10
cluster-node1-0:    serial_number: 0
cluster-node1-0: Configuration State
cluster-node1-0:    host_group: service-master
cluster-node1-0:    inventory_name: cluster-node1-0
cluster-node1-0:    inventory_vars:
cluster-node1-0:        etcd_master_addr:
cluster-node1-0:        etcd_master_name:
cluster-node1-0:        node_addr: 192.168.2.10
cluster-node1-0:        node_name: cluster-node1-0
cluster-node1-0:    ssh_address: 192.168.2.10

cluster-node2-0: Inventory State
cluster-node2-0:    name: cluster-node2-0
cluster-node2-0:    prev_state: Discovered
cluster-node2-0:    prev_status: Cancelled
cluster-node2-0:    state: Discovered
cluster-node2-0:    status: Decommissioned
cluster-node2-0: Monitoring State
cluster-node2-0:    label: cluster-node2
cluster-node2-0:    management_address: 192.168.2.11
cluster-node2-0:    serial_number: 0
cluster-node2-0: Configuration State
cluster-node2-0:    host_group: service-master
cluster-node2-0:    inventory_name: cluster-node2-0
cluster-node2-0:    inventory_vars:
cluster-node2-0:        etcd_master_addr:
cluster-node2-0:        etcd_master_name:
cluster-node2-0:        node_addr: 192.168.2.11
cluster-node2-0:        node_name: cluster-node2-0
cluster-node2-0:    ssh_address: 192.168.2.11

cluster-node3-0: Inventory State
cluster-node3-0:    name: cluster-node3-0
cluster-node3-0:    prev_state: Unknown
cluster-node3-0:    prev_status: Incomplete
cluster-node3-0:    state: Discovered
cluster-node3-0:    status: Unallocated
cluster-node3-0: Monitoring State
cluster-node3-0:    label: cluster-node3
cluster-node3-0:    management_address: 192.168.2.12
cluster-node3-0:    serial_number: 0
cluster-node3-0: Configuration State
cluster-node3-0:    host_group: service-master
cluster-node3-0:    inventory_name: cluster-node3-0
cluster-node3-0:    inventory_vars:
cluster-node3-0:        node_addr: 192.168.2.12
cluster-node3-0:        node_name: cluster-node3-0
cluster-node3-0:    ssh_address: 192.168.2.12

```

In output above the node `cluster-node1-0` is already commissioned and hence it's status is set to `Allocated`. The node `cluster-node2-0` has been decommissioned and hence it's status is set to `Decommissioned`. The node `cluster-node3-0` has not been commissioned yet, and hence it's status is set to `Unallocated`.

##### clusterctl job get last
This command dumps the ansible logs for the last configuration job that was run. It is useful to collect the output of this command in case of configuration failures.

The [Parsing Ansible Logs](#parsing-ansible-logs) section walks through a few samples of ansible logs and patterns to look for failures.

```
[vagrant@cluster-node1 ~]$ clusterctl job get last

Description: commissionEvent: nodes:[cluster-node2-0] extra-vars:{"docker_device":"/tmp/docker"} host-group:service-worker
Status: Errored
Error: exit status 2
Logs:
    [DEPRECATION WARNING]: Instead of sudo/sudo_user, use become/become_user and
    make sure become_method is 'sudo' (default).
    This feature will be removed in a
    future release. Deprecation warnings can be disabled by setting
    deprecation_warnings=False in ansible.cfg.

    PLAY [devtest] *****************************************************************
    skipping: no hosts matched

    PLAY [volplugin-test] **********************************************************
    skipping: no hosts matched

    PLAY [cluster-node] ************************************************************
    skipping: no hosts matched

    PLAY [cluster-control] *********************************************************
    skipping: no hosts matched

    PLAY [service-master] **********************************************************
    skipping: no hosts matched

    PLAY [service-worker] **********************************************************

    TASK [setup] *******************************************************************
    ok: [cluster-node2-0]

    TASK [base : include] **********************************************************
    skipping: [cluster-node2-0]

    TASK [base : include] **********************************************************
    included: /vagrant/vendor/ansible/roles/base/tasks/redhat_tasks.yml for cluster-node2-0

    TASK [base : install epel release package (redhat)] ****************************
    ok: [cluster-node2-0]

    TASK [base : install/upgrade base packages (redhat)] ***************************
    skipping: [cluster-node2-0] => (item=[u'yum-utils', u'ntp', u'unzip', u'bzip2', u'curl', u'python-requests', u'bash-completion', u'kernel', u'libselinux-python'])

...
...
```

##### clusterctl job get active
This command dumps the ansible logs for the active configuration job that is in progress. It can be useful to check the status of a running job.

```
[vagrant@cluster-node1 ~]$ clusterctl job get active

Description: decommissionEvent: nodes:[cluster-node2-0] extra-vars: {}
Status: Running
Error:
Logs:
    [DEPRECATION WARNING]: Instead of sudo/sudo_user, use become/become_user and
    make sure become_method is 'sudo' (default).
    This feature will be removed in a
    future release. Deprecation warnings can be disabled by setting
    deprecation_warnings=False in ansible.cfg.

    PLAY [all] *********************************************************************

    TASK [setup] *******************************************************************
    ok: [cluster-node2-0]

    TASK [include_vars] ************************************************************
    ok: [cluster-node2-0] => (item=contiv_network)
    ok: [cluster-node2-0] => (item=contiv_storage)
    ok: [cluster-node2-0] => (item=swarm)
    ok: [cluster-node2-0] => (item=ucp)
    ok: [cluster-node2-0] => (item=docker)
    ok: [cluster-node2-0] => (item=etcd)

    TASK [include] *****************************************************************
    included: /vagrant/vendor/ansible/roles/contiv_network/tasks/cleanup.yml for cluster-node2-0
    included: /vagrant/vendor/ansible/roles/contiv_storage/tasks/cleanup.yml for cluster-node2-0
    included: /vagrant/vendor/ansible/roles/swarm/tasks/cleanup.yml for cluster-node2-0
    included: /vagrant/vendor/ansible/roles/ucp/tasks/cleanup.yml for cluster-node2-0
    included: /vagrant/vendor/ansible/roles/docker/tasks/cleanup.yml for cluster-node2-0
    included: /vagrant/vendor/ansible/roles/etcd/tasks/cleanup.yml for cluster-node2-0
    included: /vagrant/vendor/ansible/roles/ucarp/tasks/cleanup.yml for cluster-node2-0

    TASK [stop netmaster] **********************************************************
    changed: [cluster-node2-0]

    TASK [stop aci-gw container] ***************************************************
    fatal: [cluster-node2-0]: FAILED! => {"changed": false, "failed": true, "msg": "systemd could not find the requested service \"'aci-gw'\": "}
    ...ignoring

    TASK [stop netplugin] **********************************************************

...
...
```

#### Parsing Ansible Logs
This section lists a few common patterns to look for while parsing the ansible logs gathered from the `clusterctl job get <active|last>` commands above. A common troubleshooting step is to identify the first failing task while *parsing the logs from top to down*.

##### successful task execution
The following sample logs indicates a successful execution of a task, note the status is either `changed` or `ok`.

```
    TASK [contiv_cluster : copy conf files for clusterm] ***************************
    changed: [cluster-node1] => (item=clusterm.args)
    changed: [cluster-node1] => (item=clusterm.conf)
```

```
    TASK [base : install and start ntp] ********************************************
    ok: [cluster-node2-0]
```

##### errored task execution
The following sample logs indicate the tasks have errored, note the status is `fatal`. Usually the output will also indicate the cause for failure.

```
    TASK [ucp : download and install ucp images] ***********************************
    fatal: [cluster-node2-0]: FAILED! => {"changed": true, "cmd": "docker run --rm -t --name ucp -v /var/run/docker.sock:/var/run/docker.sock docker/ucp images --image-version=1.1.0", "delt
a": "0:00:10.694920", "end": "2016-05-25 21:32:54.185655", "failed": true, "rc": 1, "start": "2016-05-25 21:32:43.490735", "stderr": "Unable to find image 'docker/ucp:latest' locally\nlates
t: Pulling from docker/ucp\ne3f56cd84e02: Pulling fs layer\n8bfd2fd2949a: Pulling fs layer\n6dc0f52613dd: Pulling fs layer\n9775bf111871: Pulling fs layer\n9775bf111871: Verifying Checksum\
n9775bf111871: Download complete\ne3f56cd84e02: Verifying Checksum\ne3f56cd84e02: Download complete\n8bfd2fd2949a: Verifying Checksum\n8bfd2fd2949a: Download complete\n6dc0f52613dd: Verifyi
ng Checksum\n6dc0f52613dd: Download complete\ne3f56cd84e02: Pull complete\n8bfd2fd2949a: Pull complete\n6dc0f52613dd: Pull complete\n9775bf111871: Pull complete\nDigest: sha256:8e28e9024127
ac0f100adbfae0f74311201dfbc8d45968e761cc5af962abf581\nStatus: Downloaded newer image for docker/ucp:latest", "stdout": "\u001b[31mFATA\u001b[0m[0000] Your engine version 1.9.1 is too old.
UCP requires at least version 1.10.0. ", "stdout_lines": ["\u001b[31mFATA\u001b[0m[0000] Your engine version 1.9.1 is too old.  UCP requires at least version 1.10.0. "], "warnings": []}
```

In output above logs the error is evident in the logs `Your engine version 1.9.1 is too old.  UCP requires at least version 1.10.0.`

```
    TASK [docker : ensure docker is started] *********************************,
    fatal: [Docker-1-FLM19379EUC]: FAILED! => {changed: false, failed: true, msg: Job for docker.service failed because the control process exited with error code. See systemctl status docker.service and journalctl -xe for details.},
```

In output above logs the failure reason is not evident in the logs but there is a hint to run subsequent commands to get the service logs using systemd commands. Also see [Systemd Service Logs](#systemd-service-logs) section for more information on these commands.

##### errored but ignored task execution
The following log indicates the task errored but the error is ignored, note the status is `fatal` but at the end `...ignoring` indicates that error was ignored.

```
    TASK [docker : restart docker (first time)] ************************************
    fatal: [cluster-node2-0]: FAILED! => {"failed": true, "msg": "The conditional check 'thin_provisioned|changed' failed. The error was: |changed expects a dictionary\n\nThe error appears
to have been in '/vagrant/vendor/ansible/roles/docker/tasks/main.yml': line 99, column 3, but may\nbe elsewhere in the file depending on the exact syntax problem.\n\nThe offending line appe
ars to be:\n\n#      of some docker bug I've not investigated.\n- name: restart docker (first time)\n  ^ here\n"}
    ...ignoring

```

#### Systemd Service Logs
All the infrastructure services are started as systemd units. When a service fails or is not behaving as desired collecting the output of following two commands is useful:

- `systemctl status <service-name>`
- `sudo journalctl -xu <service-name>`

The `<service-name>` is the name of a systemd unit. It is usually same as the name of the daemon like docker, netplugin, netmaster, volplugin, volmaster and clusterm to name a few.
