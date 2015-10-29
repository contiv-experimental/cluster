# -*- mode: ruby -*-
# vi: set ft=ruby :

gobin_dir="/opt/gopath/bin"

VAGRANTFILE_API_VERSION = "2"
Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
    config.vm.box = "contiv/centos71-netplugin"
    config.vm.box_version = "0.2.3"
    #config.vm.box = "contiv/centos71-netplugin/custom"
    #config.vm.box_url = "https://cisco.box.com/shared/static/v91yrddriwhlbq7mbkgsbbdottu5bafj.box"
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
    host_env = { }
    if ENV['CONTIV_ENV'] then
        ENV['CONTIV_ENV'].split(" ").each do |env|
            e = env.split("=")
            host_env[e[0]]=e[1]
        end
    end
    puts "Host environment: #{host_env}"
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
            # The first vm stimulates the first manually **configured** nodes
            # in a cluster
            if n == 0 then
                # mount vagrant directory such that symbolic links are copied
                #config.vm.synced_folder ".", "/vagrant", type: "rsync", rsync__args: ["--verbose", "-rLptgoD", "--delete", "-z"]
                # mount the host's gobin path for cluster related binaries to be available
                node.vm.synced_folder "#{ENV['GOPATH']}/bin", gobin_dir
                node.vm.provision "ansible" do |ansible|
                    ansible.groups = {
                        "cluster-control" => [node_name]
                    }
                    ansible.playbook = "./vendor/configuration/ansible/site.yml"
                    ansible.extra_vars = {
                        env: host_env
                    }
                end
                # expose collins port to host for ease of management
                node.vm.network "forwarded_port", guest: 9000, host: 9000
            end
        end
    end
end
