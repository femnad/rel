# rel

A simple GitHub release creator opinionated for single static binary releases.

## Assumptions

* Release is for a Go or Rust app
* There is only one binary named after the repo name
* The final binary is compiled statically
* The architecture is x86-64
* Target for Rust binaries is glibc
