#!/bin/bash

set -e

# Configure the elrepo repository so we can upgrade the kernel to 3.13
# Any kernel above 3.8 is more docker-friendly
echo "==> Configuring elrepo repository"
yum remove -y kernel kernel-headers kernel-devel
yum install -y http://www.elrepo.org/elrepo-release-7.0-2.el7.elrepo.noarch.rpm

echo "==> Installing kernel-ml from elrepo"
yum --enablerepo=elrepo-kernel install -y kernel-ml kernel-ml-headers kernel-ml-devel gcc glibc-devel glibc-headers

# Grub tweaks - default to latest kernel
echo "==> Changing grub to boot to latest kernel"
sed -i 's/^GRUB_DEFAULT=.*$/GRUB_DEFAULT=0/' /etc/default/grub
echo "==> Enabling memory limit & control of swap"
sed -ri 's/^(GRUB_CMDLINE_LINUX=.*)"$/\1 cgroup_enabled=memory swapaccount=1"/' /etc/default/grub

echo "==> Disabling selinux"
sed -i s/SELINUX=enforcing/SELINUX=disabled/g /etc/selinux/config

grub2-mkconfig >/boot/grub2/grub.cfg

# reboot
echo "Rebooting the machine..."
reboot
sleep 60
