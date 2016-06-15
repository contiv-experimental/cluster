# -*- mode: ruby -*-
# vi: set ft=ruby :

gopath_dir="/opt/gopath"
gobin_dir="#{gopath_dir}/bin"
base_ip = "192.168.2."

num_nodes = 1
if ENV['CONTIV_NODES'] && ENV['CONTIV_NODES'] != "" then
    num_nodes = ENV['CONTIV_NODES'].to_i
end

service_init = false
if ENV['CONTIV_SRV_INIT'] then
    # in demo mode we initialize and bring up the services
    service_init = true
end

clusterm_dev = true
if ENV['CONTIV_REL_CLUSTERM'] then
    # use released version of cluterm
    clusterm_dev = false
end

box = "contiv/centos72"
if ENV['CONTIV_BOX'] then
    box = ENV['CONTIV_BOX']
end

box_version = "0.6.0"
if ENV['CONTIV_BOX_VERSION'] then
    box_version = ENV['CONTIV_BOX_VERSION']
end

host_env = { }
if ENV['CONTIV_ENV'] then
    ENV['CONTIV_ENV'].split(" ").each do |env|
        e = env.split("=")
        host_env[e[0]]=e[1]
    end
end

if ENV["http_proxy"]
  host_env["HTTP_PROXY"]  = host_env["http_proxy"]  = ENV["http_proxy"]
  host_env["HTTPS_PROXY"] = host_env["https_proxy"] = ENV["https_proxy"]
  host_env["NO_PROXY"]    = host_env["no_proxy"]    = ENV["no_proxy"]
end

ceph_vars = {
    "fsid" => "4a158d27-f750-41d5-9e7f-26ce4c9d2d45",
    "monitor_secret" => "AQAWqilTCDh7CBAAawXt6kyTgLFCxSvJhTEmuw==",
    "journal_size" => 100,
    "monitor_interface" => "eth1",
    "cluster_network" => "#{base_ip}0/24",
    "public_network" => "#{base_ip}0/24",
    "journal_collocation" => "true",
    "devices" => [ '/dev/sdb', '/dev/sdc' ],
}

ansible_groups = { }
bootstrap_node_ansible_groups = { }
ansible_playbook = ENV["CONTIV_ANSIBLE_PLAYBOOK"] || "./vendor/ansible/site.yml"
ansible_extra_vars = {
    "env" => host_env,
    "service_vip" => "#{base_ip}252",
    "validate_certs" => "no",
    "control_interface" => "eth1",
    "netplugin_if" => "eth2",
    "docker_version" => "1.10.3",
    "scheduler_provider" => "native-swarm",
}
ansible_extra_vars = ansible_extra_vars.merge(ceph_vars)

shell_provision = <<EOF
#give write permission to go path directory to be able to run tests
chown -R vagrant:vagrant #{gopath_dir}

# if we are coming up in non-demo environment then load
# the clusterm binaries from dev workspace
if [ "#{clusterm_dev}" = "true" ]; then
    echo mounting binaries from dev workspace
    rm -f /usr/bin/clusterm /usr/bin/clusterctl
    ln -s #{gopath_dir}/src/github.com/contiv/cluster/management/src/bin/clusterm /usr/bin/clusterm
    ln -s #{gopath_dir}/src/github.com/contiv/cluster/management/src/bin/clusterctl /usr/bin/clusterctl
else
    echo using released binaries
fi
EOF

VAGRANTFILE_API_VERSION = "2"
Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
    config.vm.box = box
    config.vm.box_version = box_version
    node_ips = num_nodes.times.collect { |n| base_ip + "#{n+10}" }
    node_names = num_nodes.times.collect { |n| "cluster-node#{n+1}" }
    # this is to avoid the issue: https://github.com/mitchellh/vagrant/issues/5186
    config.ssh.insert_key = false
    # use a private key from within the repo for demo environment. This is used for
    # pushing configuration
    config.ssh.private_key_path = "./management/src/demo/files/insecure_private_key"
    (0..num_nodes-1).reverse_each do |n|
        node_name = node_names[n]
        node_addr = node_ips[n]
        node_vars = {
            "etcd_master_addr" => node_ips[0],
            "etcd_master_name" => node_names[0],
            "ucp_bootstrap_node_name" => node_names[0],
        }
        config.vm.define node_name do |node|
            node.vm.hostname = node_name
            # create an interface for cluster (control) traffic
            node.vm.network :private_network, ip: node_addr, virtualbox__intnet: "true"
            # create an interface for cluster (data) traffic
            node.vm.network :private_network, ip: "0.0.0.0", virtualbox__intnet: "true"
            node.vm.provider "virtualbox" do |v|
                # give enough ram and memory for docker to run fine
                v.customize ['modifyvm', :id, '--memory', "4096"]
                v.customize ["modifyvm", :id, "--cpus", "2"]
                # make all nics 'virtio' to take benefit of builtin vlan tag
                # support, which otherwise needs to be enabled in Intel drivers,
                # which are used by default by virtualbox
                v.customize ['modifyvm', :id, '--nictype1', 'virtio']
                v.customize ['modifyvm', :id, '--nictype2', 'virtio']
                v.customize ['modifyvm', :id, '--nictype3', 'virtio']
                v.customize ['modifyvm', :id, '--nicpromisc2', 'allow-all']
                v.customize ['modifyvm', :id, '--nicpromisc3', 'allow-all']
                # XXX: creating disk doesn't work in stock centos box, remove this check
                # once we need ceph working in stock OS demo
                if box == "contiv/centos72" then
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
            end

            # in dev mode, provision base packages needed for cluster management
            # by tests
            if clusterm_dev then
                if ansible_groups["cluster-node"] == nil then
                    ansible_groups["cluster-node"] = [ ]
                end
                ansible_groups["cluster-node"] << node_name
            end

            # The first vm stimulates the first manually **configured** nodes
            # in a cluster
            if n == 0 then
                # mount vagrant directory such that symbolic links are copied
                #node.vm.synced_folder ".", "/vagrant", type: "rsync", rsync__args: ["--verbose", "-rLptgoD", "--delete", "-z"]
                #mount the repo directory to vm's go-path for running system-tests
                node.vm.synced_folder ".", "#{gopath_dir}/src/github.com/contiv/cluster"
                #mount the repo directory to a fixed directory, that get's referred
                #in a test specific conf file (see the test-suite setup function)
                node.vm.synced_folder ".", "/vagrant"

                # expose collins port to host for ease of management
                node.vm.network "forwarded_port", guest: 9000, host: 9000

                # expose UCP UI to the host on port
                node.vm.network "forwarded_port", guest: 443, host:9091 

                # add this node to cluster-control host group
                ansible_groups["cluster-control"] = [node_name]
            end

            if service_init then
                # Share anything in `shared` to '/shared' on the cluster hosts.
                node.vm.synced_folder "shared", "/shared"

                ansible_extra_vars = ansible_extra_vars.merge(node_vars)
                # first 3 nodes are master, out of that first 2 nodes are for bootstrap
                if n < 3 then
                    # for bootstrap-node we need to use a separate host group variable
                    # as otherwise `vagrant provision` ends up running on all hosts.
                    # This seems to be due to difference in provisioning behavior
                    # between `vagrant up` and `vagrant provision`
                    if n < 2 then
                        if bootstrap_node_ansible_groups["service-master"] == nil then
                            bootstrap_node_ansible_groups["service-master"] = [ ]
                        end
                        bootstrap_node_ansible_groups["service-master"] << node_name
                    end
                    # if we are bringing up services as part of the cluster, then start
                    # master services on the first three vms
                    if ansible_groups["service-master"] == nil then
                        ansible_groups["service-master"] = [ ]
                    end
                    ansible_groups["service-master"] << node_name
                else
                    # if we are bringing up services as part of the cluster, then start
                    # worker services on rest of the vms
                    if ansible_groups["service-worker"] == nil then
                        ansible_groups["service-worker"] = [ ]
                    end
                    ansible_groups["service-worker"] << node_name
                end
            end

            # Run the provisioners after all machines are up
            if n == 0 then
                node.vm.provision 'bootstrap-ansible', type: 'ansible' do |ansible|
                    ansible.groups = bootstrap_node_ansible_groups
                    ansible.playbook = ansible_playbook
                    ansible.extra_vars = ansible_extra_vars.clone
                    ansible.limit = 'all'
                end
                node.vm.provision 'main-ansible', type: 'ansible' do |ansible|
                    ansible.groups = ansible_groups
                    ansible.playbook = ansible_playbook
                    ansible.extra_vars = ansible_extra_vars.clone
                    # Turn off init cluster as this is a member-add scenario
                    ansible.extra_vars["etcd_init_cluster"] = false
                    ansible.limit = 'all'
                end
                # run the shell provisioner on the first node, after all the provisioners,
                # to correctly mount dev binaries if needed
                node.vm.provision 'devenv-shell', type: "shell" do |s|
                    s.inline = shell_provision
                end
            end
        end
    end
end
