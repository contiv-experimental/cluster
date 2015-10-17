# -*- mode: ruby -*-
# vi: set ft=ruby :

gobin_dir="/opt/gopath/bin"

# provision script for common packages
provision_common = <<SCRIPT
## setup the environment file. Export the env-vars passed as args to 'vagrant up'
echo Args passed: [[ $@ ]]
echo > /etc/profile.d/envvar.sh
echo PATH=$PATH:#{gobin_dir} >> /etc/profile.d/envvar.sh
if [ $# -gt 0 ]; then
    echo "export $@" >> /etc/profile.d/envvar.sh
fi

source /etc/profile.d/envvar.sh
SCRIPT

# provision script for control vm specific packages
provision_control = <<SCRIPT
## pass the env-var args to docker. This helps passing stuff like http-proxy etc
if [ $# -gt 0 ]; then
    (mkdir /usr/lib/systemd/system/docker.service.d) || exit 1
    (echo [Service] > /usr/lib/systemd/system/docker.service.d/http-proxy.conf) || exit 1
    (IFS=' '; for env in $@; \
        do echo Environment="$env" >> /usr/lib/systemd/system/docker.service.d/http-proxy.conf; \
        done) || exit 1
fi

## start docker service
(systemctl daemon-reload && service docker restart) || exit 1

## download and start collins container
(docker run -dit -p 9000:9000 --name collins tumblr/collins) || exit 1

## start cluster manager
(echo starting cluster manager)
(sleep 60 && clusterm 0<&- &>/tmp/clusterm.log &) || exit 1
SCRIPT


VAGRANTFILE_API_VERSION = "2"
Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
    #config.vm.box = "contiv/centos71-netplugin"
    #config.vm.box_version = "0.2.2"
    config.vm.box = "contiv/centos71-netplugin/custom"
    config.vm.box_url = "https://cisco.box.com/shared/static/v91yrddriwhlbq7mbkgsbbdottu5bafj.box"
    num_nodes = 1
    if ENV['CONTIV_NODES'] && ENV['CONTIV_NODES'] != "" then
        num_nodes = ENV['CONTIV_NODES'].to_i
    end
    base_ip = "192.168.2."
    node_ips = num_nodes.times.collect { |n| base_ip + "#{n+10}" }
    node_names = num_nodes.times.collect { |n| "cluster-node#{n+1}" }
    # this is to avoid the issue: https://github.com/mitchellh/vagrant/issues/5186
    config.ssh.insert_key = false
    # use a private key from within the repo for demo environment. This is used for
    # pushing configuration
    config.ssh.private_key_path = "./management/src/demo/files/insecure_private_key"
    num_nodes.times do |n|
        node_name = node_names[n]
        node_addr = node_ips[n]
        config.vm.define node_name do |node|
            node.vm.hostname = node_name
            # create an interface for cluster (control) traffic
            node.vm.network :private_network, ip: node_addr, virtualbox__intnet: "true"
            node.vm.provider "virtualbox" do |v|
                # make all nics 'virtio' to take benefit of builtin vlan tag
                # support, which otherwise needs to be enabled in Intel drivers,
                # which are used by default by virtualbox
                v.customize ['modifyvm', :id, '--nictype1', 'virtio']
                v.customize ['modifyvm', :id, '--nictype2', 'virtio']
                v.customize ['modifyvm', :id, '--nicpromisc2', 'allow-all']
            end
            node.vm.provision "shell" do |s|
                s.inline = provision_common
                s.args = ENV['CONTIV_ENV']
            end
            # The first vm stimulates the first manually **configured** nodes
            # in a cluster
            if n == 0 then
                # mount vagrant directory such that symbolic links are copied
                #config.vm.synced_folder ".", "/vagrant", type: "rsync", rsync__args: ["--verbose", "-rLptgoD", "--delete", "-z"]
                # mount the host's gobin path for cluster related binaries to be available
                node.vm.synced_folder "#{ENV['GOPATH']}/bin", gobin_dir
                node.vm.provision "shell" do |s|
                    s.inline = provision_control
                    s.args = ENV['CONTIV_ENV']
                end
                # expose collins port to host for ease of management
                node.vm.network "forwarded_port", guest: 9000, host: 9000
            end
provision_node = <<SCRIPT
## set hostname ourselves, somehow vagran't hostname config doesn't
## work for the first vm
(echo setting hostname)
(hostnamectl set-hostname #{node_name}) || exit 1

## install necessary iptables to let mdns work
(echo setting up iptables for mdns)
(iptables -I INPUT -p udp --dport 5353 -i eth1  -j ACCEPT && \
 iptables -I INPUT -p udp --sport 5353 -i eth1  -j ACCEPT) || exit 1

## start serf
(echo starting serf)
(serf agent -discover mycluster -iface eth1 \
 -tag NodeLabel=`hostname` \
 -tag NodeSerial=`lshw -c system | grep serial | awk '{print $2}'` \
 -tag NodeAddr=#{node_addr} \
 0<&- &>/tmp/serf.log &) || exit 1
SCRIPT
            node.vm.provision "shell" do |s|
                s.inline = provision_node
            end
        end
    end
end
