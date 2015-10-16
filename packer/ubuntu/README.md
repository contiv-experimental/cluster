## Ubuntu Images

Adapted from https://github.com/boxcutter/ubuntu

You can read more about this project
[here](https://github.com/contiv/build/blob/master/packer/README.md). There are
additional requirements in this README you should be aware of.

To build this project, and start a vm with the new box, type `make`.

Other useful `make` targets:
*  `make ssh` will ssh into the started VM. It will also start the VM with
   vagrant if necessary.
* targets `start` and `stop` will start and destroy the VM, respectively.
