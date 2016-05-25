# Contributing to Cluster

Patches and feature additions are totally welcome to cluster project! 

Interested in contributing code? Please see the [Developers section](#developers) to setup a build and test environment.

There are other ways to contribute, not just code and feature additions: report issues, propose documentation changes, submit bug fixes, propose design changes, discuss use cases or become a maintainer.

## Reporting issues

To report issues that pertain to code behavior, the following command outputs should
be included in the issue description:
- `clusterctl --version`
- `clusterm --version`

Following are some other useful command outputs to include for reporting issues related to failures in discovery, commission or decommision:
- `clusterctl job get last`
- `clusterctl job get active`
- `clusterctl nodes get`

Also refer the [troubleshooting guide](./management/troubleshoot.md) for more details on common steps to follow for narrowing down on failures.

## Submitting pull requests to change the documentation or the code

Changes can be proposed by sending a pull request (PR). A maintainer
will review the changes and provide feedback.

The pull request will be merged into the master branch after discussion.

Please make sure to run the tests and that the tests pass before
submitting the PR. Please keep in mind that some changes might not be
merged if the maintainers decide they can't be merged.

Please squash your commits to one commit per fix or feature. The resulting
commit should have a single meaningful message.

## Commit message guidelines

The short summary should include the name of the directory or file affected by
the commit (e.g.: `add boltdb inventory subsystem`).

A longer description of what the commit does should start on the third
line when such a description is deemed necessary.

If you have trouble with the appropriate git commands to handle these
requirements, please let us know! We're happy to help.

## Legal Stuff: Sign your work
You must sign off on your work by adding your signature at the end of the
commit message. Your signature certifies that you wrote the patch or
otherwise have the right to pass it on as an open-source patch.
By signing off your work you ascertain following (from [developercertificate.org](http://developercertificate.org/)):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
660 York Street, Suite 102,
San Francisco, CA 94110 USA

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.

Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

Every git commit message must have the following at the end on a separate line:

    Signed-off-by: Joe Smith <joe.smith@email.com>

Your real legal name has to be used. Anonymous contributions or contributions
submitted using pseudonyms cannot be accepted.

Two examples of commit messages with the sign-off message can be found below:
```
clusterm: fix bug

This fixes a random bug encountered in clusterm.

Signed-off-by: Joe Smith <joe.smith@email.com>
```
```
clusterm: fix bug

Signed-off-by: Joe Smith <joe.smith@email.com>
```

If you set your `user.name` and `user.email` git configuration options, you can
sign your commits automatically with `git commit -s`.

These git options can be set using the following commands:
```
git config user.name "Joe Smith"
git config user.email joe.smith@email.com
```

`git commit -s` should be used now to sign the commits automatically, instead of
`git commit`.

## Developers

### Environment Requirement

- vagrant 1.7.3 or higher
- virtualbox 5.0 or higher
- ansible 2.0 or higher
- docker 1.10 or higher

**Note**: At the moment, development environment is only supported for Linux based systems.

### Fork `contiv/cluster` Repo and Clone

```
mkdir ./workspace 
cd ./workspace
git clone https://github.com/<userid>/cluster.git
```

### Build binaries

```
cd ./workspace/cluster/management/src/
make build
```

### Run Unit Tests
```
cd ./workspace/cluster/management/src/
make unit-test
```

### Run System Tests
```
cd ./workspace/cluster/management/src/
make system-test 
```
