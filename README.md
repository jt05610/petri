# Petri

Petri is a full-stack framework for designing and automating lab research processes. It is designed for mid-level
automation: automation of lab processes that would be benefited by automation, but are not done with enough volume to
justify a full automation system.

Petri is designed to be easy to use, and easy to extend. It has a [user interface](https://github.com/jt05610/petri-ui),
code generator,
and [gRPC API](https://github.com/jt05610/otto).

## Rationale


## Install

## Usage

### How to automate your research with Petri

#### Define the components within the system

Individual automata within Petri are defined with elementary [Petri Nets](https://en.wikipedia.org/wiki/Petri_net).
Petri Nets
were chosen over finite state machines for system modeling because they are easily composable. Petri Nets have been used
to model chemical processes for decades, and are a natural fit for modeling lab processes.

While Petri Nets could theoretically be used to model everything that happens down to the molecular level, we have found
it most useful to start to model components at the highest level possible (i.e. the places and transitions that you want
to affect other components, rather than every possible thing something can do).

The components in the [standard library](https://github.com/jt05610/petri-std) are a good place to see examples of the
components used to automate the author's research.

#### Generate the code

Currently, code generation in Go is fully supported, and Python is almost there.

#### Implement the code

Petri currently supports the following serial protocols for communicating with devices:

- [x] [Modbus](https://en.wikipedia.org/wiki/Modbus) (support up to 255 devices on one bus)
- [x] [Marlin](https://marlinfw.org/) (Used to make 3D printers do lab things)
- [x] [grbl](https://github.com/grbl/grbl) (Used to quickly make stepper motors do things using a GRBL shield and an
  arduino, supports up to 4 motors.)

Network protocols:

- [x] [gRPC](https://grpc.io/) (https://github.com/jt05610/petri-proto)
- [x] [GraphQL](https://graphql.org/) (https://github.com/jt05610/petri-graph)

#### Run the code

Run your generated servers and the Petri Daemon will automatically discover them and connect them to the network. This
is done using simple heartbeating over [RabbitMQ](https://www.rabbitmq.com/).

#### Sequence events with the UI

You can create a sequence of events using the Petri UI. The Petri Net acts like a contract between devices, so they will
always know how to play nicely together once defined.

#### Run your system

Single executions are performed with the UI. Batch executions are currently performed using the
GraphQL API and the [otto python client](https://www.github.com/jt05610/otto) which incorporates Design of Experiments
algorithms, Batch Bayesian Optimization, and a CouchDB database to manage samples, batches, results, analyses, and
optimization experiments.

