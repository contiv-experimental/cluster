##Preparing baremetal or personal VM environments for Cluster Management

This document goes through the steps to prepare a baremetal or personal VM setup for cluster management.

**Note:**
- Unless explicitly mentioned all the steps below are done by logging into the same host. It is referred as *control host* below.
- Right now we test cluster manager on Centos7.2. More OS variations shall be added in future.

###0. ensure following pre-requisites are met on the control host
- ansible 2.0 or higher is installed.
- git is installed.
- a management user has been created. Let's call that user **cluster-admin** from now on.
  - note that `cluster-admin` can be an existing user.
  - this user needs to have **passwordless sudo** access configured. You can use `visudo` tool for this.
- a ssh key has been generated for `cluster-admin`. You can use `ssh-keygen` tool for this.
- the public key for `cluster-admin` user is added to all other hosts in your setup. You can use `ssh-copy-id cluster-admin@<hostname>` for this, where `<hostname>` is name of the host in your setup where `cluster-admin` is being added as authorized user.

###1. download and install cluster manager on the control host
```
# Login as `cluster-admin` user before running following commands
git clone https://github.com/contiv/ansible.git
cd ansible

# Create inventory file
echo [cluster-control] > /tmp/hosts
echo node1 ansible_host=127.0.0.1 >> /tmp/hosts

# Install cluster manager
ansible-playbook --key-file=~/.ssh/id_rsa -i /tmp/hosts -e '{"env": {}}' ./site.yml
```

**Note**:
- `env` is a mandatory variable. It is used to specify the environment for running ansible tasks like http-proxy. If there is no special environment to be setup then it needs to be set to an empty dictionary as shown in the example above.

###2. setup cluster manager configuration on the control host
Edit the cluster manager configuration file that is created at `/etc/default/clusterm/clusterm.conf` to setup the user and playbook-location information. A sample is shown below. `playbook-location` needs to be set as the path of ansible directory we cloned in previous step. `user` needs to be set as name of `cluster-admin` user and `priv_key_file` is the location of the `id_rsa` file of `cluster-admin` user.
```
# cat /etc/default/clusterm/clusterm.conf
{
    "ansible": {
        "playbook-location": "/home/cluster-admin/ansible/",
        "user": "cluster-admin",
        "priv_key_file": "/home/cluster-admin/.ssh/id_rsa"
    }
}
```
After the changes look good, restart cluster manager
```
sudo systemctl restart clusterm
```

###3. provision rest of the nodes for discovery from the control host
Cluster manager uses serf as discovery service for node-health monitoring and ease of management. Here we will provision all the hosts to be added to the discovery service. The command takes `<host-ip>` as an argument. This is the IP address of the interface (also referred as control-interface) connected to the subnet designated for carrying traffic of infrastructure services like serf, etcd, swarm etc
```
clusterctl discover <host-ip>
```

**Note**:
- Once the above command is run for a host, it shall start showing up in `clusterctl nodes get` output in a few minutes.
- the `clusterctl discover` command expects `env` and `control_interface` ansible variables to be specified. This can be achieved by using the `--extra-vars` flag or setting the variables at global level using `clusterctl global set` command. For more information on other available variables, also checkout [discovery section of ansible vars](ansible_vars.md#serf-based-discovery)

###4. ready to rock and roll!
All set now, you can follow the cluster manager workflows as [described here](./README.md#get-list-of-discovered-nodes).
