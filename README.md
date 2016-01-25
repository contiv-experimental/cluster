[![Build Status](http://contiv.ngrok.io/view/Cluster%20CI/job/Cluster%20Push%20Build%20Master/badge/icon)](http://contiv.ngrok.io/view/Cluster%20CI/job/Cluster%20Push%20Build%20Master/)

##Overview
Contiv provides an integrated stack of components to launch containerized applications in
a multi-tenant environment. The stack offers following features:
- Application definition, launch and monitoring
- Built-in service discovery
- [storage](https://github.com/contiv/volplugin) and [network](https://github.com/contiv/netplugin)
  provisioning for user and infra apps
- Cluster management
  - [Inventory](#inventory)
    - discovery
    - node state database
  - [Node image management](#node-image-management)
    - install
    - upgrade
  - [Node configuration management](#node-configuration-management)
    - third party packages (ovs, serf, consul/etcd, docker etc)
    - infrastructure services (netmaster/netplugin and volmaster/volplugin)
  - [Bootstrap](#bootstrap-1)
    - node
    - cluster
  - [Lifecycle](#lifecycle-1)
    - node
    - cluster

The rest of the the document details the **cluster management** requirements of the stack
that were briefly listed above. The design choices to meet these requirements are captured
[here.](management/DESIGN.md)

##Definitions and Terms:
###Server/Node/Host/Asset
A standalone physical/virtual machine with predefined/discovered compute, memory, storage
and networking capabilities.

###Cluster
A collection of one or more nodes. Nodes may or may not be homogeneous in their capabilities.

###Base Image/Image
A base OS image with *minimal set* of pre-installed packages that are needed for cluster
management service to manage that node. For instance, for automated discovery of node a
cluster membership service like Serf needs to be pre-installed in the base OS and started
at node's bootup.

###Bootstrap
When used in context of node, it refers to the process of installing the node image and
bringing up the node first time. It is sometime also referred to as node onboarding. 

When used in context of cluster, bootstrap refers to the process of adding first bootstrapped
node to the cluster.

###Lifecycle
A collection of well-defined states that a node or cluster goes through when certain events happen.

###Infra App/Infra Service
An userspace application or a kernel module that is used to provide one or more system services
like plugins/drivers or controller components for networking and storage; key/value stores; etc.

##Inventory
Inventory management requires:
- **Cluster membership management**: It should be possible to automatically discover the nodes
and track their discovery status.
- **Node state database**: It should be possible to track the current state of the node in
the cluster. A few possible states are node has been discoverd but waiting to be
commissioned; node has been commissioned; node is being upgraded; and node has
been decommissioned.

##Node Image Management
Node image management requires: 
- **Image repository**: It should be possible to have the base image available at a central
repository. The image can then be made available to be booted using a system like pxe.
- **Image installation**: It should be possible to automate image installation on new nodes
with minimal user involvement as provided by kickstart.
- **Image upgrade**: It should be possible to trigger cluster wide or rolling image upgrades
from a central location.

**Note:** To start simple this feature is **not** provided as part of initial
release of cluster manager. In the initial release the operator shall be provided the base image
as a bootable iso file that they will need to integrate with their node onborading process.

##Node Configuration Management
Node configuration management requires:
- **Configuration repository**: It should be possible to have the configuration for deploying the
services available at a central repository. The configuration can then be made available to be
deployed on appropriate nodes.
- **Configuration push**: It should be possible to automatically push the configuration to the
nodes when they are commissioned. It should be possible to tweak some of the configuration to
fit user's environment.
- **Configuration cleanup**: It should be possible to automatically cleanup the configuration
from the nodes when they are decommissioned.
- **Configuration verification**: It should be possible to verify the configuration deployed on
the nodes.
- **Configuration upgrade**: It should be possible to trigger cluster wide or rolling configuration
upgrades from a central location.
- **Role/Group based configuration**: It should be possible to selectively start services on nodes
based on their group or role. It should be possible for the node to be assigned to roles or groups
statically by the operator or dynamically based on service availabilty policy.

**Note:** To start with dynamic role assignment may not be supported in initial release of
cluster manager.

##Bootstrap
Bootstraping involves bringing up a node or cluster for the first time.

###Node
The node bootstrap requires:
- installing the base image
- performing first time configurations like disk partitioning
- allocating an IP for node to become reachable to rest of the cluster management.
- setting up user and credentials with enough permissions to perfom configuration management.
- starting the pre-requisite services like Serf
 
**Note:** To start simple this feature is **not** provided as part of initial
release of cluster manager. In the initial release the user shall be provided the base image
as a *bootable iso* file that they will need to integrate with their node onborading process.

###Cluster
The cluster bootstrap requires adding the first bootstrapped node to the cluster by starting
the cluster management service on it with appropriate parameters like:
- configuration management parameters like user information, configuration repository
- inventory management paramters like database url

**Note:** To start simple this feature is provided in form of a bootstrap script that the
operator will need to run on the first node inorder to bring up the cluster management service. 

##Lifecycle
Lifecycle management involves integration of various cluster management aspects at node and
cluster level to a central place for ease of monitoring and administration by the cluster operator. 
 
###Node
Node lifecycle management requires handling the following:
- **Bootstrap**: It should be possible to bootstrap the nodes remotely. *As described above this
feature is not provided in initial release of the cluster manager.*
- **Cluster membership**: It should be possible to automatically track cluster's membership of
reachable and unreachable nodes.
- **Commission**: It should be possible for the operator to remotely commission a reachable node.
Optionally it should be possible to automatically commission newly discovered node. This shall be
configurable in cluster manager configuration parameter.
- **Upgrade**: It should be possible for the operator to remotely upgrade a reachable node.
Optionally it should be possible to automatically upgrade nodes when configuration repository
changes. This shall be configurable in cluster manager configuration parameter.
- **Decommission**: It should be possible for the operator to remotely decommission a node.
- **Batch operations**: It should be possble for the operator to commission, upgrade or decommission
all or a batch of nodes.
- **Reloads**: In event of node reloads (or node reachabilty changes) it should be possible to verify the
node's configuration and automate corrective actions in case of failures.

###Cluster
- **Bootstrap**: It should be possible to bootstrap the cluster remotely. *As described above this
feature is provided in form a bootstrap script that will need to be manually run on the first node
in initial release of the cluster manager.*
- **Sustenance and Availabilty**: Cluster manager service shall remain available as long as there
is one node up and running in the cluster.
