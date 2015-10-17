ansible_provision = proc do |ansible|
  ansible.playbook = 'ansible/netplugin/site.yml'

  proxy_env = { }

  %w[http_proxy https_proxy].each do |name|
    if ENV[name]
      proxy_env[name] = ENV[name]
    end
  end

  # In a production deployment, these should be secret
  ansible.extra_vars = {
    proxy_env: proxy_env,
  }

  ansible.limit = 'all'
end

Vagrant.configure(2) do |config|
  config.ssh.insert_key = false # workaround for https://github.com/mitchellh/vagrant/issues/5048

  config.vm.define "centos" do |centos|
    centos.vm.box = "contiv/centos71-netplugin"
  end

  config.vm.define "ubuntu" do |ubuntu|
    ubuntu.vm.box = "contiv/ubuntu1504-netplugin"
    # so it only runs once
    ubuntu.vm.provision 'ansible', &ansible_provision
  end
end
