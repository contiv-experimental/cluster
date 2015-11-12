# -*- mode: ruby -*-
# vi: set ft=ruby :

gobin_dir="/opt/gopath/bin"
base_ip = "192.168.2."

num_nodes = 1
if ENV['CONTIV_NODES'] && ENV['CONTIV_NODES'] != "" then
    num_nodes = ENV['CONTIV_NODES'].to_i
end

service_init = false
if ENV['CONTIV_SRV_INIT'] then
    service_init = true
end

host_env = { }
if ENV['CONTIV_ENV'] then
    ENV['CONTIV_ENV'].split(" ").each do |env|
        e = env.split("=")
        host_env[e[0]]=e[1]
    end
end
puts "Host environment: #{host_env}"

ceph_vars = {
    "fsid" => "4a158d27-f750-41d5-9e7f-26ce4c9d2d45",
    "monitor_secret" => "AQAWqilTCDh7CBAAawXt6kyTgLFCxSvJhTEmuw==",
    "journal_size" => 100,
    "monitor_interface" => "eth1",
    "cluster_network" => "#{base_ip}0/24",
    "public_network" => "#{base_ip}0/24"
}

ansible_groups = { }
ansible_playbook = "./vendor/configuration/ansible/site.yml"
ansible_extra_vars = {
    "env" => host_env
}
ansible_extra_vars = ansible_extra_vars.merge(ceph_vars)

VAGRANTFILE_API_VERSION = "2"
Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
    config.vm.box = "contiv/centos71-netplugin"
    config.vm.box_version = "0.3.1"
    #config.vm.box = "contiv/centos71-netplugin/custom"
    #config.vm.box_url = "https://cisco.box.com/shared/static/v91yrddriwhlbq7mbkgsbbdottu5bafj.box"
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
        node_vars = {
            "online_master_addr" => node_ips[0],
        }
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
                # create disks for ceph
                (0..1).each do |d|
                  disk_path = "disk-#{n}-#{d}"
                  vdi_disk_path = disk_path + ".vdi"

                  v.customize ['createhd',
                               '--filename', disk_path,
                               '--size', '11000']
                  # Controller names are dependent on the VM being built.
                  # It is set when the base box is made in our case ubuntu/trusty64.
                  # Be careful while changing the box.
                  v.customize ['storageattach', :id,
                               '--storagectl', 'SATA Controller',
                               '--port', 3 + d,
                               '--type', 'hdd',
                               '--medium', vdi_disk_path]
                end
            end
            # The first vm stimulates the first manually **configured** nodes
            # in a cluster
            if n == 0 then
                # mount vagrant directory such that symbolic links are copied
                #node.vm.synced_folder ".", "/vagrant", type: "rsync", rsync__args: ["--verbose", "-rLptgoD", "--delete", "-z"]
                # mount the host's gobin path for cluster related binaries to be available
                node.vm.synced_folder "#{ENV['GOPATH']}/bin", gobin_dir

                # expose collins port to host for ease of management
                node.vm.network "forwarded_port", guest: 9000, host: 9000

                # add this node to cluster-control host group
                ansible_groups["cluster-control"] = [node_name]

                if service_init then
                    # if we are bringing up services as part of the cluster, then start
                    # master services on first vm
                    ansible_groups["service-master"] = [node_name]
                    ansible_groups["ceph-hosts"] = [node_name]
                    ansible_extra_vars = ansible_extra_vars.merge(node_vars)
                end
            elsif service_init then
                # if we are bringing up services as part of the cluster, then start
                # worker services on rest of the vms
                if ansible_groups["service-worker"] == nil then
                    ansible_groups["service-worker"] = [ ]
                end
                ansible_groups["service-worker"] << node_name
                ansible_groups["ceph-hosts"] << node_name
                ansible_extra_vars = ansible_extra_vars.merge(node_vars)
            end

            if service_init
                # Share anything in `shared` to '/shared' on the cluster hosts.
                node.vm.synced_folder "shared", "/shared"
            end

            # Run the provisioner after all machines are up
            if n == (num_nodes - 1) then
                node.vm.provision 'ansible' do |ansible|
                    ansible.groups = ansible_groups
                    ansible.playbook = ansible_playbook
                    ansible.extra_vars = ansible_extra_vars
                    ansible.limit = 'all'
                end
            end
        end
    end
end
