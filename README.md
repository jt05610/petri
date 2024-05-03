# Petri

Petri is a full-stack framework for designing and automating lab research processes. It is designed for mid-level
automation: automation of lab processes that would be benefited by automation, but are not done with enough volume to
justify a full automation system.

Petri is designed to be easy to use, and easy to extend. It has a [user interface](https://github.com/jt05610/petri-ui),
[code generator](https://github.com/jt05610/petri-codegen),
and [gRPC API](https://github.com/jt05610/otto).

## Rationale

### The problem

#### The automation gap

It is impossible to build an automated system that can do everything in a lab. Current lab-automation technologies and
systems are designed for high-throughput automation of specific processes such as high content screening. This works
well for processes that are done with high volume, but is not feasible for processes that are done with low volume. The
majority of lab work is therefore done by hand in most labs, and only high-value and high-volume processes are
automated. Research suffers as a result, as time is wasted to learn how to do techniques by hand and there is a risk of
human-error being introduced with every manual step.

#### The replication crisis

We are in an ongoing [Replication crisis](https://en.wikipedia.org/wiki/Replication_crisis) in science. While there are
many causes of this crisis, a root cause is a lack of standardization in the way science is done. Outside the scientific
method, there is not a framework for how to do science, and every lab has their own way of doing things. This makes it
difficult to reproduce results and build on the work of others.

### The solution

> Nobody sits around before creating a new Rails project to figure out where they want to put their views; they just run
> rails new to get a standard project skeleton like everybody else.
>
> [Cookiecutter Data Science](http://drivendata.github.io/cookiecutter-data-science/#other-people-will-thank-you)

Petri is inspired by frameworks such as Ruby on Rails, Django, and React. These frameworks facilitate software
development by handling the boilerplate and common tasks that are required for every project, allowing developers to
mainly focus on what they are developing rather than how they are developing it.

Petri is designed to do the same for lab experiments. Rather than try to automate everything, Petri is designed to
handle networking, data storage, and code boilerplate generation for modular components within scalable automated
systems. This allows researchers to put together systems for their needs.

## Install

### Docker (recommended)

The easiest way to get started with Petri is to use the [docker image](https://hub.docker.com/r/jt05610/petri).

```bash
docker pull jt05610/petri:latest
```

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

Petri currently supports the following serial protocols for communicating with your devices:

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

