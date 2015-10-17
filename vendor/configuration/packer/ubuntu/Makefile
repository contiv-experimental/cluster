all: stop build add start 

add:
	vagrant box add ubuntu1504 box/ubuntu1504-netplugin.box --force

start:
	vagrant up

stop:
	vagrant destroy -f

restart: stop start

build: download
	version=$$(cat VERSION) atlas_token=${ATLAS_TOKEN} packer build --force ubuntu1504.json

download:
	bash download-kernel.sh

# XXX this is to work around a packer issue that causes the vbox startup to
# fail every time it's run after the first time.
rmcache:
	rm -rf packer_cache

ssh: start
	vagrant ssh
