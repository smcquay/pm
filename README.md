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
platform: darwin/amd64
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
