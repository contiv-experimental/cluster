#Contiv Cluster Manager Roadmap

This documents contains the features that are mandatory for cluster manager to
be considered functionally complete. This is more of a todo list and shall be
obsoleted once we have milestone/release level feature tracking in place.

**Note:** This doesn't tracks bugs and requests for features that are good to
have. They shall be tracked through github issues.

- [ ] cluster manager to provide APIs for batched node operations
- [x] cluster manager to accept it's configuration as command line arguments
- [ ] cluster manager to allow changing some of the configuration without requiring
      a restart.
- [ ] auto commission of nodes
- [ ] cluster manager to accept values for ansible specific group-vars etc as part of
      command line
- [ ] ability to assign roles/group to commissioned nodes both statically and dynamically
- [ ] harden the node lifecycle especially to deal with failures
- [ ] ansible playbooks to provision, upgrade, cleanup, verify 
  - [ ] netmaster/netplugin service
    - [ ] host level network configuration like bundling NICs
    - [ ] ovs
  - [ ] volmaster/volplugin service
    - [ ] ceph
  - [x] etcd datastore
  - [ ] consul datastore
  - [ ] VIP service for high availability
  - [ ] docker stack including daemon, swarm and orca containers
  - [ ] what else?
