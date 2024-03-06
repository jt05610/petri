---
title: 'Petri: a systems integration platform for laboratory automation'
tags:
  - Go
  - Python
  - Protocol Buffers
  - systems integration
  - discrete event systems
  - petri nets
authors:
  - name: Jonathan R Taylor
    orcid: 0000-0000-0000-0000
    affiliation: "1" # (Multiple affiliations must be quoted)
  - name: Uday Kompella
    affiliation: "1,2,3" # (Multiple affiliations must be quoted)

affiliations:
  - name: University of Colorado Anschutz Medical Campus, Department of Pharmaceutical Sciences, University of Colorado, Aurora, CO
    index: 1
  - name: Departments of Bioengineering and Ophthalmology, University of Colorado Anschutz Medical Campus, Aurora, CO
    index: 2
  - name: Colorado Center for Nanomedicine and Nanosafety, University of Colorado Anschutz Medical Campus, Aurora, CO
    index: 3
date: 13 August 2017
bibliography: paper.bib
---

# Summary

Laboratories are composed of multiple systems that are often not designed to
work together, and the integration of these systems is a common challenge in
modern research laboratories. While significant advancements have been made in
developing sophisticated lab equipment and software, the ability to seamlessly
integrate these systems is often limited. This is particularly true in the
field of pharmaceutical sciences, where the need to integrate a wide variety of
systems, including liquid handling robots, plate readers, data analysis
scripts, and lab information management systems. Here, we present Petri, a
schema-first Go-based platform for systems integration that is designed to be
flexible, scalable, and easy to use. Petri provides a common interface to
describe systems and the relationships between components within the system.
With Petri, systems modeling and implementation are handled separately,
facilitating sharing systems and workflows between labs with completely
different equipment and software services, or allowing easy changing of one
service or piece of equipment with another.

# Statement of need

`Petri` is a Go-based service for integrating laboratory systems with each
other. Go enables 'Petri' to be fast, scalable, and with memory safety. The API
was designed to be schema-first, so that the focus while designing is on what
systems need to do, rather than the details of how they will do. This allows
great flexibility when implementing and sharing systems.  `Petri` has a code
generator that to generate code for components in the systems, and a runtime
that can be used to run the systems. `Petri` relies heavily on abstractions and
interfaces, so that it can be used with any kind of lab instrument or software
service, independent of the component's provided interface. A GraphQL API is
provided to interact with the runtime with a scripting language of the user's
choice, and a web-based interface is provided to design systems and sequences
of events.

`Petri` was designed to facilitate integration of DIY and commercially
available components with a variety of interfaces. It has been used in the
author's thesis work to integrate DIY and commercially available components to
automate lipid nanoparticle library preparation and characterization. The
ability to rapidly design and integrate systems with `Petri` will enable
researchers to develop efficient, reproducible, and sharable workflows that are
resistant to changes in lab members, equipment, and software services.

As its name suggests, `Petri` is based on the concept of a petri net, which are
used to model discrete event systems. While there are many petri net software
packages available, `Petri` is unique in that it is designed to be used designing
and running distributed systems, rather than modeling, simulation, and analysis
of petri nets.

# Systems integration with Petri

`Petri` is based on the concept of a petri net, a mathematical model of
discrete event systems. A petri net is a directed bipartite graph, consisting of
places, transitions, and arcs. An example of a petri net is:
![Switch-light-logger system petri net](pump_net.svg)

where the ellipses are places, the rectangles are transitions, and the arrows
are arcs. The places contain tokens, and the transitions are enabled when the
places leading to them contain tokens.

`Petri` uses colored petri nets, where there are different types of tokens.
Token schema are generated, and the arcs are labeled with the token schema that
is extracted from the place and given to the transition.

## Controlling transitions

Transitions are fired as soon as they are active. Often, this is not the desired behavior. There are many Petri Net
concepts that have been created for handling this. `Petri` provides two such mechanisms:

1. The transition can have a "guard" that is a function that is called to determine if the transition is enabled.
   If the guard returns false, the transition is not enabled even if the tokens are present.

2. The transition can have an `Event` placed on it. In this case, the transition will only fire when the event is
   called and all required tokens are present. This is described with petri using a `.yaml` file.

## Doing real work

Once systems are defined with petri nets, code is generated to implement the
system in the language of the user's choice. Protocol Buffers are used to
serialize data between services in the system. To communicate, the services
use RabbitMQ, a message broker that implements the Advanced Message Queuing
Protocol (AMQP). This allows the services to be distributed across multiple
machines, and to be written in any language that has a RabbitMQ client.

To interface with RabbitMQ, `Petri` maps the places of the petri net to
RabbitMQ queues. Events get mapped to queues that share the same name as the
transition the event is mapped to. This allows communication between services,
as each device is both a producer and a consumer of messages.

# Availability and future directions

`Petri` is available on GitHub. A docker-compose is provided to run the system with
Docker. The project is licensed under the MIT license.

A simple tutorial is provided to demonstrate how to use `Petri` to quickly design and run systems using python and Go.

# Acknowledgements

