##Overview
Contiv provides an integrated stack of components to launch containerized applications in
a multi-tenant environment. The stack offers following features:
- application definition, launch and monitoring
- built-in service discovery
- storage and network provisioning for user and infra apps
- fleet/cluster management
   - inventory
   - node image management
     -  upgrades
   - node configuration management
     - base OS components (ovs, serf, consul/etcd etc)
     - infra apps (netmaster, netplugin, volmaster, volplugin etc)
   - bootstrap
     - single node
     - cluster
   - node lifecycle
     - birth
     - configure and infra bringup
     - discovery and cluster membership
     - reloads
     - upgrades
   - cluster lifecycle
     - init
     - role assignment (TBD: it might be a bit early to address this, but leving it here for comments)
     - node add, remove, upgrade

The rest of the the document details the fleet/cluster management aspects of the stack.

##Definitions and Terms:
###Server/node/host
A standalone physical/virtual machine with predefined/discovered compute, memory, storage and networking capabilities.

###Cluster
A collection of one or more nodes. Nodes may or may not be homogeneous in their capabilities.

###Node image/image
A base OS image with minimal set of pre-installed packages

###Bootstrap
When used in context of node, it refers to process of booting up the node first time by one of the options:
 - getting an IP addres for one of it's interfaces (possibily using DHCP) connected to management network
   and downloading a new image (using pxe boot)
 - or using a pre configured IP address and pre installed image
 - or a mix of both.
 
When used in context of cluster, it refers to process of adding first node to the cluster.

###Lifecycle
A collection of states that a node or cluster goes through when certain events happen.

###Infra app
An userspace application or a kernel module that is used to provide one or more system services like plugins/drivers or controller components for networking and storage; key/value stores; etc.

##Inventory
Inventory management involves:
 - Provisioning a new node: A few defining criteria could be node's dns name; ip address; (TBD: certificates/serial-numbers/security tokens? for verification and node discovery)
 - Deprovisioning a node

Available choices for inventory management:
 - [ ] ansible file based inventory
   - might be basic and not provide features like node verification
 - [ ] external database/REST URI + ansible 
 - [ ] home grown UI, database and node verification
 - [ ] ??

##Node image management
Node image management involves:
 - central image storage or a pxe boot server for downloading images
 - tracking per node image (in case of heterogeneous nodes wrt image)
 - triggering cluster wide or rolling image upgrades

Available choices for node image management:
 - [ ] home grown UI, image database and management logic
 - [ ] ??

##Node configuration management
Node configuration management involves:
 - central repository of infra apps (like rpm repository) for downlaoding packages
 - upgrading one or multiple services

**Note** that separating out configuration management from image management obviates the need for replacing node image and perform (often disruptive) node reloads when only certain system services need to be upgraded.

Available choices for node configuration management:
 - [ ] ansible playbooks 
 - [ ] home grown UI and configuration database
 - [ ] ??

##Bootstrap
Bootstraping involves bringing up a node or cluster for the first time.

###Node
Based on deployment, this may need to be handled as a special case than say node reloads if a node needs to be booted with a pxe image and configured for infra apps.

Available choices for node bootstrap management:
 - [ ] pxe boot + dhcp IPAM from a dedicated server
 - [ ] ??

###Cluster
TBD

##Lifecycle
Lifecycle management involves integration of various cluster management aspects (node and/or cluster level) to a central place for ease of monitoring and administration by the cluster operator. 
 
Available choices for node/cluster bootstrap management:
 - [ ] home grown UI and backend(s) (like serf for discovery; ansible for configuration; dhcp and pxe for node bootstrap)
 - [ ] ??
