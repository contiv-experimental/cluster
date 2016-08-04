## 3 steps to Contiv Cluster Management

If you are trying cluster manager with baremetal hosts or a personal VM setup follow [this link](./baremetal.md) to setup the hosts. After that you can manage the cluster as described in the [step 3.](#3-login-to-the-first-node-to-manage-the-cluster) below.

To try with a built-in vagrant based VM environment continue reading.

### 0. Ensure correct dependencies are installed
- vagrant 1.7.3 or higher
- virtualbox 5.0 or higher
- ansible 2.0 or higher

### 1. checkout and build the code
```
cd $GOPATH/src/github.com/contiv/
git clone https://github.com/contiv/cluster.git
```

### 2. launch three vagrant nodes. 

**Note:** If you look at the project's `Vagrantfile`, you will notice that all the vagrant nodes (except for the first node) boot up with stock centos7.2 OS and a `serf` agent running. `serf` is used as the node discovery service. This is intentional to meet the goal of limiting the amount of services that user needs to setup to start bringing up a cluster and hence making management easier.
```
cd cluster/
CONTIV_NODES=3 make demo-cluster
```

### 3. login to the first node to manage the cluster

**Note:** The first node is slightly special in a way that it is booted up with two additional services viz. `collins` and `clusterm`. `collins` is used as the node lifecycle management and event logging service. `clusterm` is the cluster manager daemon. `clusterctl` utility is provided to exercise cluster manager provided REST endpoint.

```
CONTIV_NODES=3 vagrant ssh cluster-node1
```

#### Provision additional nodes for discovery
```
clusterctl discover <host-ip(s)>
```
Cluster Manager uses Serf as a discovery service for node health monitoring and for cluster bootstrapping. Use `discover` command to include additional nodes in the discovery service. The `<host-ip>` should be an IP address from a management network only used by infra services such as serf, etcd, swarm, etc..

**Note**:
```
clusterctl discover 192.168.2.11 192.168.2.12 --extra-vars='{"env" : {}, "control_interface": "eth1" }'
```
- The command above will provision the other two vms (viz. cluster-node2 and cluster-node3) in the vagrant setup for serf based discovery. Once it is run, the discovered hosts will appear in `clusterctl nodes get` output in a few minutes.
- the `clusterctl discover` command expects `env` and `control_interface` ansible variables to be specified. This can be achieved by using the `--extra-vars` flag as shown above or by setting them at [global level](#setget-global-variables), if applicable. For more information on other available variables, also checkout [discovery section of ansible vars](ansible_vars.md#serf-based-discovery)

#### Get list of discovered nodes
```
clusterctl nodes get
```

And info for a single node can be fetched by using `clusterctl node get <node-name>`.

#### Commission a node
```
clusterctl node commission <node-name> --host-group=<service-master|service-worker>
```

Commissioning a node involves pushing the configuration and starting infra services on that node using `ansible` based configuration management. The services that are configured depend on the mandatory parameter `--host-group`. Checkout the `service-master` and `service-worker` host-groups in [ansible/site.yml](../vendor/ansible/site.yml) to learn more about the services that are configured. To quickly check if commissioning a node worked, you can run `etcdctl member list` on the node. It shall list all the commissioned members in the list.

**Note**:
- certain ansible variables need to be set for provisioning a node. The list of mandatory and other useful variables is provided in [ansible_vars.md](./ansible_vars.md). The variables need to be passed as a quoted JSON string in node commission command using the `--extra-vars` flag.
```
clusterctl node commission node1 --extra-vars='{"env" : {}, "control_interface": "eth1", "netplugin_if": "eth2" }' --host-group "service-master"
```
- a common set of variables (like environment) can be set just once as [global variables](#setget-global-variables). This eliminates the need to specify the common variables for every commission command.

#### Decommission a node
```
clusterctl node decommission <node-name>
```

Decommissioning a node involves stopping and cleaning the configuration for infra services on that node using `ansible` based configuration management.

#### Update a node
```
clusterctl node update <node-name>
```

Updating a node involves updating the configuration for infra services on that node using `ansible` based configuration management. Other use-cases for updating a node include installing newer versions of infra services or changing the host-group of the node like changing a node from worker to master and vice-versa.

**Note**:
```
clusterctl node update node1 --extra-vars='{"env" : {}, "control_interface": "eth1", "netplugin_if": "eth2" }' --host-group "service-worker"
```
- similar to [commission](#commission-a-node) command, the `--extra-vars` flag can be used with the `update` command to specify ansible variables needed for provisioning the node.
- to change the host-group of a node, the `--host-group` flag is used. If this flag is not specified then node's configuration is updated with the last set host-group.

#### Set/Get global variables
```
clusterctl global set --extra-vars=<vars>
```
A common set of variables (like environment, scheduler-provider and so on) can be set just once using the `--extra-vars` flag with `clusterctl global set` command.

**Note**:
- The variables need to be passed as a quoted JSON string using the `--extra-vars` flag.
```
clusterctl global set --extra-vars='{"env" : {"http_proxy": "my.proxy.url"}, "scheduler_provider": "ucp-swarm"}'
```
- The variables set at global level are merged with the variables specified at the node level, with the latter taking precedence in case of an overlap/conflict.
- The list of useful variables is provided in [ansible_vars.md](./ansible_vars.md).

#### Get provisioning job status
```
clusterctl job get <active|last>
```
Common cluster management workflows like commission, decommission and so on involve running an ansible playbook. Each such run per workflow is referred to as a job. You can see the status of an ongoing (active) or last run job using this command.

#### Managing multiple nodes
```
clusterctl nodes commission <space separated node-name(s)>
clusterctl nodes decommission <space separated node-name(s)>
clusterctl nodes update <space separated node-name(s)>
```

The worflow to commission, decommission or update all or a subset of nodes can be performed by using `clusterctl nodes` subcommands. Please refer the documentation of individual commands above for details.

##Want to learn more?
Read the [design spec](DESIGN.md) and/or see the remaining/upcoming features in [github issues page](https://github.com/contiv/cluster/issues)
