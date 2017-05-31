# Demo Controller

## Goal

The goal of this project is to demonstrate how you can build a simple Kubernetes controller.

This is not meant as a project to be used directly - but rather as a reference point to build your own custom controllers.

This example is currently based off client-go v3.0.0-beta.0 - but will be updated as new versions become available.

## Helpful Resources

- Upstream controller development and design principles
    - https://github.com/kubernetes/community/blob/master/contributors/devel/controllers.md
    - https://github.com/kubernetes/community/blob/master/contributors/design-proposals/principles.md#control-logic

- Upstream Kubernetes controller package
    - https://github.com/kubernetes/kubernetes/tree/release-1.6/pkg/controller

- client-go examples (version sensitive, e.g. use v3 examples with v3 checkout)
    - https://github.com/kubernetes/client-go/tree/v3.0.0-beta.0/examples

- Creating Kubernetes Operators Presentation (@metral)
    - http://bit.ly/lax-k8s-operator

- Memcached operator written in python (@pst)
    - https://github.com/kbst/memcached

## Roadmap

- Demonstrate using
    - leader-election
    - Third Party Resources
    - Shared Informers
    - Events

## Building

Build agent and controller binaries:

`make clean all`

Build agent and controller Docker images:

`make clean images`

