## 3 steps to Contiv Cluster Management

If you are trying cluster manager with baremetal hosts or a personal VM setup follow [this link](./baremetal.md) to setup the hosts. After that you can manage the cluster as described in the [step 3.](#3-login-to-the-first-node-to-manage-the-cluster) below.

To try with a built-in vagrant based VM environment continue reading.

### 0. Ensure correct dependencies are installed
- docker 1.9 or higher
- vagrant 1.7.3 or higher
- virtualbox 5.0 or higher
- ansible 2.0 or higher

### 1. checkout and build the code
```
cd $GOPATH/src/github.com/contiv/
git clone https://github.com/contiv/cluster.git
cd cluster/management/src
make build
```

**Note**: `build` is run inside a docker container, so it requires docker to be installed
and running on the host. To avoid having to use `sudo` for builds you can add the current
user to `docker` group once by using the following command:
```
sudo usermod -a -G docker `id -un`
```

### 2. launch three vagrant nodes. 

**Note:** If you look at the project's `Vagrantfile`, you will notice that all the vagrant nodes (except for the first node) boot up with stock centos7.2 OS and a `serf` agent running. `serf` is used as the node discovery service. This is intentional to meet the goal of limiting the amount of services that user needs to setup to start bringing up a cluster and hence making management easier.
```
cd ../..
CONTIV_NODES=3 vagrant up
```

### 3. login to the first node to manage the cluster

**Note:** The first node is slightly special in a way that it is booted up with two additional services viz. `collins` and `clusterm`. `collins` is used as the node lifecycle management and event logging service. `clusterm` is the cluster manager daemon. `clusterctl` utility is provided to exercise cluster manager provided REST endpoint.

```
CONTIV_NODES=3 vagrant ssh cluster-node1
```

#### Get list of discovered nodes
```
clusterctl nodes get
```

And info for a single node can be fetched by using `clusterctl node get <node-name>`.

#### Commision a node
```
clusterctl node commission <node-name>
```

Commissioning a node involves pushing the configuration and starting infra service on that node using `ansible` based configuration management. Checkout the `service-master` and `service-worker` groups in [ansible/site.yml](../vendor/configuration/ansible/site.yml) to find out more about the serivces that configured. To quickly check if commisioning a node worked, you can run `etcdctl member list` on the node. It shall list all the commissioned members in the list.

**Note**:
- certain ansible variables  need to be set for provisioning a node. The list of mandatory and other useful variables is provided in [ansible_vars.md](./ansible_vars.md). The variables need to be passed as a quoted JSON string in node commission command using the `--extra-vars` flag.
```
clusterctl node commission node1 --extra-vars='{"env" : {"http_proxy": "my.proxy.url"}, "control_if": "eth2", "netplugin_if": "eth1" }'
```
- a common set of variables (like environment info etc) can be set just once using `clusterctl global set` command with `--extra-vars` flag set.
- The variables set at global level are merged with the variables specified as per node level, with latter taking precendence in case of an overlap/conflict.

#### Decommision a node
```
clusterctl node decommission <node-name>
```

Decommissioning a node involves stopping and cleaning the configuration for infra services on that node using `ansible` based configuration management.

#### Perform an upgrade
```
clusterctl node maintain <node-name>
```

Upgrading a node involves upgrading the configuration for infra services on that node using `ansible` based configuration management.

#### Managing multiple nodes

**To be added**. This shall allow commission, decommission and rolling-upgrades of all or a subset of nodes.

##Want to learn more?
Read the [design spec](DESIGN.md) and/or see the remaining/upcoming items in [roadmap](ROADMAP.md)
