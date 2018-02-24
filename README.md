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
