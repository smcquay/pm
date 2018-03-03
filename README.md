# pm: a simple, cross-platform system package manager

`pm` exists amid a set of trade-offs in distributing software. The ideas behind
`pm` were born at a time when:

- There was no overlap in the Venn diagram of system package managers that
  offered both strong security promises (signed packages) and permissive
  licensing (most are GPL).
- There was reason to suspect that Unix systems might be shipped without
  scripting languages; software like [brew](https://brew.sh) would cease to
  work and engineers would be left without a way to fetch and install software.
- Engineers wanted to deploy software to a variety of Unix-like environments
  using a single system.
- Engineers wanted a simple-to-reason-about system that used familiar Unix
  primitives as building blocks to distribute their software.

Simplicity is a principal design goal of this project. When offered an
opportunity to chose between two designs the design that requires less mental
scaffolding to describe or implement should be used. As a concrete example:
transitive dependency calculations are implemented, but supporting compatible
version *ranges* are not.

The project is currently in early design phases, and this document describes
the high-level approach of the project.

## Components

There are two main components to this project.

0. `pm` is the name of the client-side cli command. This is the tool used to
   fetch, install, verify, create, upload, etc. packages.
0. `pmd` is the name of the server-side component. It hosts packages (over
   `http` for now), available package metadata, and cryptographic public key
   information to clients.

Securely installing the `pm` command is important. Be sure to verify its
contents before use.

## Package Format

The intention is to be able to create and open package files with commonly used
Unix utilities. The package file is an uncompressed
[tar](https://en.wikipedia.org/wiki/Tar_(computing)) file contaning the
following files:

0. `meta.yaml` -- contains information about the package's contents, and is
   transmitted to clients during for which available packages a remote can
   serve, e.g.:
```yaml
name: foo
version: 2.3.29
description: Foo is the world's simplest frobnicator
deps: [baz, bar@0.9.2]
```

0. `root.tar.bz2` -- A compressed tarball that will eventually be expanded
   starting at `$PM_ROOT`
0. `bom.sha256` -- [checksum](https://s.mcquay.me/sm/cs) file containing sha256
   checksums of the expected contents of `root.tar.bz2`
0. `manifest.sha256` -- [checksum](https://s.mcquay.me/sm/cs) file of the
   expected contents of the `.pkg` file.
0. `manifest.sha256.asc` -- [OpenPGP](https://www.openpgp.org) detached
   signature for the `manifest.sha256` file. Its validity communicates that the
   contents have not been tampered with.
0. `bin/{pre,post}-{install,ugrade,remove}` (**optional**) -- a collection of
   executables that are run at the relevant stages.

As a minimum package authors are required to author the `root.tar.bz2` and the
`meta.yaml` files, and the `pm pkg create` will generate the rest of the files,
using the key information associated with the `PM_PGP_EMAIL` environment
variable.
If you can make a [tar file](https://en.wikipedia.org/wiki/Tar_(computing)) and write
a [yaml](http://yaml.org) file, you can create a `pm`package! 


## Remote Repositories

The notion of `remote` is borrowed from [git](https://git-scm.com); a `pm`
client can be configured to pull packages from multiple remote repositories. It
is intended to be trivial to deploy `pmd`, and equally trivial to configure
clients to fetch from multiple `remote`s.

The example remote url:

`https://pm.mcquay.me/darwin/amd64/testing`

encodes a remote that is served over `https` on the host `pm.mcquay.me` and
informs the client to pull packages from the `/darwin/amd64/testing` namespace,
specified by the Path. `pm pull` will collect available package information
from configured remote and will populate its local database with the contents
of the response. `pm` can then list available packages, and the user can then
request that they be installed.

As a practical example a client can be configured to pull from two `remotes`
as:

```bash
$ pm add remote https://pm.mcquay.me/darwin/amd64/stable
$ pm add remote https://pm.example.com/generic/testing
$ pm pull
$ pm available
foo     0.1.2      https://pm.mcquay.me/darwin/amd64/stable
bar     3.2.3      https://pm.mcquay.me/generic/testing
```

Here each remote advertises one package each. After pulling metadata from the
`remote` server the client database is populated, and the user listed all
installable packages. In the case of collisions the first configured `remote`
offering a colliding packages will be the used.

Previous versions of `pm` use to implicitly formulate namespace values based on
host information (os and arch), but allowing package maintainers and end users
to specify this value explicitly allows for greater flexibility. 
