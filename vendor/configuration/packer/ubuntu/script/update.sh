#!/bin/bash -eux

set -e

# Disable the release upgrader
echo "==> Disabling the release upgrader"
sed -i.bak 's/^Prompt=.*$/Prompt=never/' /etc/update-manager/release-upgrades

echo "==> Updating list of repositories"
# apt-get update does not actually perform updates, it just downloads and indexes the list of packages
apt-get -y update

echo "==> Performing dist-upgrade (all packages and kernel)"
apt-get -y dist-upgrade --force-yes

dpkg -i /tmp/*.deb

echo "==> Rebooting the vm"
reboot
sleep 60
