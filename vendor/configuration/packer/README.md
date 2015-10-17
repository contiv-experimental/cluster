## Packer Configurations

Configurations for `centos` and `ubuntu` are available. To build them
simultaneously, build with `make`. Alternatively, `build-#` targets are
available, which will build each one independently (replace `#` with the os
type).

## Notes on configuration

* Release docker (latest version) is installed on all hosts
* Care was taken to ensure kernel versions are similar. 4.1 is currently used
  as of this writing. This is to ensure we are compatible with all features of
  docker.
* You will want ansible on your host. Install with `pip install ansible`,
  typically as root.

## Requirements

* Packer 0.8.2
* VirtualBox 5.0+ (needed to support vbox tools under 4.x kernel)
* Vagrant 1.7.4+ (to support VBox 5.0+)
* Ansible 1.9.2
