#Contiv Cluster Manager Design

This document provides a specification of the core design elements of the contiv
cluster manager. The design is influenced from the cluster management requirements
as identified in the [primary README.md](../README.md). A few premises for this design are:
- Make use of and learn from existing systems wherever applicable. And the
most prominent evenidence of this can be seen in the fact that the design depends
heavily on and takes benefit of some of the well known and deployed open source
systems viz. [Serf](https://www.serfdom.io/), [Collins](http://tumblr.github.io/collins/index.html)
and [Ansible](http://www.ansible.com/) to accompalish the cluster management requirements.
- provide an easy to use and programmable interface to the cluster adminstrator, while
  not loosing/hiding the functionality of underlying systems.

**Note:** The node image management requirement is not addressed in the first release
of cluster manager. However, care is taken in the design to make it non disruptive to
add this feature later by adding/using the states provided by the node lifecycle management system.

The document is organized into following sections:

- [Subsystems](#subsystems)
  - [Node Inventory](#node-inventory)
    - [Collins](#collins)
    - [Node Lifecycle](#node-lifecycle)
  - [Node Monitoring](#node-monitoring)
    - [Serf](#serf)
  - [Node Configuration](#node-configuration)
    - [Ansible](#ansible)
    - [Provisioning](#provisioning)
    - [Cleanup](#cleanup)
    - [Upgrade](#upgrade)
    - [Verification](#verification)
- [Manager](#manager)
  - [Configuration](#configuration)
  - [REST interface](#rest-interface)
  - [Events and Event Loop](#events-and-event-loop)
  - [Cluster Lifecycle](#cluster-lifecycle)

##Subsystems
Cluster manager consists of three subsystems viz. inventory, monitoring and configuration.
These susbsystems correspond to all-of/part-of the function provided by the open source
systems viz. collins, serf and ansible respectively. Defining these subsystem boundaries
clearly keeps the design modular and provides well defined interface that the core cluster
manager can use.

These subsystems are described in sections below.

###Node Inventory
Inventory subsystem provides the following:
- a database of nodes and their respective lifecycle states
- a logging of node related events like state changes, failures etc

####Collins
Collins is an open source inventory system that provides a rich set of APIs for
managing node lifecycle among other things. You can read more about [Collins here](http://tumblr.github.io/collins/index.html)

####Node Lifecycle
Collins supports a well defined set of [node lifecycle status'](http://tumblr.github.io/collins/concepts.html#status%20&%20state). 

Following is description of lifecycle transitions as implemented in cluster manager.
- **First time discovery**: When a node is discovered it is moved to `Unallocated` status with state `Discovered`. There are only two possible states of a node viz. `Discovered` and `Disappeared`. They represent the current status of the node as reported by the monitoring system.
- **Commission a node**: When a node is commisioned by the user it is first moved to `Provisioning` status. In this status the configuration is pushed to the node using Ansible configuration management subsystem. This is where the services are deployed on the node. Once the provisioning completes the node is moved to `Allocated` status. In event of configuration failure the node is moved back to `Unallocated` status
- **Decommision a node**: When a node is decommisioned by the user it is first moved to `Cancelled` status. In this status the configuration is cleanup from the node using Ansible configuration management subsystem. This is where the services are stopped on the node. Once the cleanup completes the node is moved to `Decommissioned` status.
- **Upgrade a node**: When a node is upgraded by the user it is first moved to `Maintenance` status. In this status the new configuration is pushed to the node using Ansible configuration management subsystem. This is where the services are upgrade on the node. Once the upgrade completes the node is moved back to `Allocated` status. In event of configuration failure the node is moved to `Unallocated` status.

**Note:** Along with node status transitions the result of configuration push is updated there as well. [**TBD**: the logging of configuration events need to be done.]

###Node Monitoring
Monitoring subsystem provides the following:
- a mechanism to monitor and distribute node's reachability in the cluster management plane.
- a mechanism to distribute node specific information like label, serial-number and management IP,
  that is used by the inventory and configuration subsystem for their function.

####Serf
Serf is an open source system for cluster membership and failure detection. You can read more about [Serf here](https://www.serfdom.io/).

###Node Configuration
Configuration subsystem provides the following:
- a mechanism to push, upgrade, cleanup and verify configuration on a node based on it's role

####Ansible
Ansible is a open source system for configuration management. You can read more about [Ansible here](http://www.ansible.com/). In particular we use the ansible playbooks to deploy services. The following sub-sections describe the different types of playbooks that cluster manager uses for managing the services.

**Note:** Since there will be more than one service that we will deploy in our cluster, we need a way to organize the playbooks such that they can be tested independently (in respective service workspace) while we are able to invoke them through a single playbook that includes them.

####Provisioning
A playbook to provision a service perfoms the various actions needed to configure and run that service. This playbook is run when a node is commisioned.

####Cleanup
A playbook to cleanup a service performs the various actions needed to stop and remove that service. This playbook is run when a node is decommisioned.

####Upgrade
A playbook to upgrade a service performs the various actions needed to update the configuration and restart that service. This playbook is run when a node is upgraded.

####Verification
A playbook to verify a service performs the various actions needed to verify status of a service. This playbook is run when a node is commissioned or upgraded. [**TBD**: may be this should be part of the above playbooks themselves?]

##Manager
Cluster manager drives the node lifecycle by listening to `monitor` subsystem and `user` events. Cluster manager provides REST endpoints for user driven events like commisioning, decommissioning and maintaining/upgrading a node.

**Note:** As of now cluster manager runs only on one node (i.e. the first node). In a real deployment, running cluster manager will be tied to node life-cycles and it will be a highly available service.

###Configuration
[**TBD**: add the configuration details here]

###REST interface
[**TBD**: add the REST interface spec here]

###Events and Event Loop
Cluster manager is an event based system. An event may correspond to a trigger from one of the subsystems like node getting discovered. An event can also be user triggered like commissioning a new node. And processing an event might generate more events like commisioning a node puts it in `Provisioning` status and triggers configuration event which pushes configuration to the node and puts the node in appropriate state based on configuration result.

Cluster manager runs an event loop that processes one event at a time before moving to next event. The events are processed in the order in which they are enqueued. An event processing may acquire locks on affected nodes inorder to serialize node acesses by different conflicting events.

**TBD**: the locking facility needs to be implemented.
**TBD**: add details on events and respective processing

###Cluster Lifecycle
The cluster lifecycle consists of two stages:
- **bootstrap**: This is the stage where the first node in the cluster is brought up and needs to be manually configured to start the cluster manager service.
- **sustenance**: This is the stage where one or more nodes are active in a cluster at any time. In this stage the cluster manager service is always available and shall be used to perform and node related operations.
