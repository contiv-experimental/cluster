#!/bin/bash -eux

echo '==> Configuring sshd_config options'

echo '==> Turning off sshd DNS lookup to prevent timeout delay'
echo "UseDNS no" >> /etc/ssh/sshd_config
echo '==> Disabling GSSAPI authentication to prevent timeout delay'
echo "GSSAPIAuthentication no" >> /etc/ssh/sshd_config
echo "==> Setting MaxSessions to 1000. Needed for systemtests"
echo "MaxSessions 1000" >> /etc/ssh/sshd_config
