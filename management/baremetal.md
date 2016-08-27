##Baremetal Server or Virtual Machine Installation

This document goes through the steps to install Contiv Cluster to a baremetal
or virtual machine.

**Note:**
- Unless explicitly mentioned all the steps below are done by logging into the same host. It is referred as *control host* below.
- Right now we test cluster manager on Centos7.2. More OS variations will be added in future.

###0. Ensure the following pre-requisites are met on the control host
- ansible 2.0 or higher is installed.
- git is installed.
- a management user has been created. Let's call that user **cluster-admin** from now on.
  - note that `cluster-admin` can be an existing user.
  - this user needs to have **passwordless sudo** access configured. You can use `visudo` tool for this.
- a ssh key has been generated for `cluster-admin`. You can use `ssh-keygen` tool for this.
- the public key for `cluster-admin` user is added to all the hosts(including the control host) in your setup. You can use `ssh-copy-id cluster-admin@<hostname>` for this, where `<hostname>` is name of the host in your setup where `cluster-admin` is being added as authorized user.

###1. Download and install Cluster Manager
```
# Login as `cluster-admin` user before running following commands
git clone https://github.com/contiv/ansible.git
cd ansible

# Create an inventory file
echo [cluster-control] > /tmp/hosts
echo node1 ansible_host=127.0.0.1 >> /tmp/hosts

# Install Cluster Manager service
ansible-playbook --key-file=~/.ssh/id_rsa -i /tmp/hosts -e '{"env": {}, "control_interface": "ifname"}' ./site.yml
```

**Note**:
- `env` and `control_interface` need to be specified.
- `env` is used to specify the environment for running ansible tasks like http-proxy. If there is no special environment to be setup then it needs to be set to an empty dictionary as shown in the example above.
- `control_interface` is the netdevice that will carry serf traffic on this node.

###2. Configure the cluster manager service
Edit the cluster manager configuration file that is created at `/etc/default/clusterm/clusterm.conf` to setup the user and playbook-location information. A sample is shown below. `playbook-location` needs to be set as the path of ansible directory we cloned in previous step. `user` needs to be set as name of `cluster-admin` user and `priv_key_file` is the location of the `id_rsa` file of `cluster-admin` user.
```
# cat /etc/default/clusterm/clusterm.conf
{
    "ansible": {
        "playbook_location": "/home/cluster-admin/ansible/",
        "user": "cluster-admin",
        "priv_key_file": "/home/cluster-admin/.ssh/id_rsa"
    }
}
```
After the changes look good, signal cluster manager to load the updated configuration
```
sudo systemctl kill -sHUP clusterm
```

###3. Ready to rock and roll!
All set now, you can follow the cluster manager workflows as described [here](./README.md#provision-additional-nodes-for-discovery).
