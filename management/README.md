## 3 steps to Contiv Cluster Management

### 1. checkout and build the code
```
cd $GOPATH/src/github.com/contiv/
git clone https://github.com/mapuri/cluster.git
cd cluster/management/src
git checkout collins
make build
```

### 2. launch three vagrant nodes. 

**Note:** If you look at the project's `Vagrantfile`, you will notice that all the vagrant nodes (except for the first node) boot up with stock centos7.1 os and a `serf` agent running. `serf` is used as the node discovery service. This is intentional to meet the goal of limiting the amount of services that user needs to setup to start bringing up a cluster and hence making management easier.
```
cd ..
CONTIV_ENV="http_proxy=$http_proxy https_proxy=$https_proxy" CONTIV_NODES=3 vagrant up
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

Commissioning a node involves pushing the configuration and starting infra service on that node using `ansible` based configuration management. For now a simple playbook `src/demo/files/site.yml` is used that creates the file `/tmp/yay` on commissioned node.

#### Decommision a node
```
clusterctl node decommission <node-name>
```

Decommissioning a node involves stopping and cleaning the configuration for infra services on that node using `ansible` based configuration management.  For now a simple playbook `src/demo/files/cleanup.yml` is used that deletes the file `/tmp/yay` from decommissioned node.

#### Perform an upgrade
```
clusterctl node maintain <node-name>
```

Upgrading a node involves upgrading the configuration for infra services on that node using `ansible` based configuration management. For now a simple playbook `src/demo/files/ansible.yml` is used that touches the file `/tmp/yay` on node being maintined.

#### Managing multiple nodes

**To be added**. This shall allow commission, decommission and rolling-upgrades of all or a subset of nodes.

##Want to learn more?
Read the [design spec](DESIGN.md) and/or see the upcoming features in [roadmap](ROADMAP.md)
