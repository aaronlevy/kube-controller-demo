# Demo Controller

# Goal

The goal of this project is to demonstrate how you can build a simple Kubernetes controller.

This is not meant as a project to be used directly - but rather as a reference point to build your own custom controllers.

This example is currently based off client-go v2.0.0 - but will be updated as new versions become available.

# Roadmap

- Update to client-go v3.0.0 (when available)
- Demonstrate using leader-election
- Demonstrate using work-queues
- Demonstrate using Third Party Resources
- Demonstrate using Shared Informers

# Building

Build agent and controller binaries:

`make clean all`

Build agent and controller Docker images:

`make clean images`

