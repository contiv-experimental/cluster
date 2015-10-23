#!/bin/bash -eux

echo "UseDNS no" >> /etc/ssh/sshd_config
echo "MaxSessions 1000" >> /etc/ssh/sshd_config
