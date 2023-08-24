from typing import Callable

from valve.domain import commands, events
from valve.service_layer import unit_of_work


def open_a(
        cmd: commands.OpenA,
        uow: unit_of_work.AbstractUnitOfWork,
):
    with uow:
        uow.do(cmd)


def open_b(
        cmd: commands.OpenB,
        uow: unit_of_work.AbstractUnitOfWork,
):
    with uow:
        uow.do(cmd)


def flow(
        cmd: commands.Flow,
        uow: unit_of_work.AbstractUnitOfWork,
):
    with uow:
        uow.do(cmd)


def publish_a_opened(
        event: events.AOpened,
        publish: Callable
):
    publish(events.AOpened)


def publish_b_opened(
        event: events.BOpened,
        publish: Callable
):
    publish(event)


def publish_flow_a(
        event: events.FlowedA,
        publish: Callable
):
    publish(event)


def publish_flow_b(
        event: events.FlowedB,
        publish: Callable
):
    publish(event)


def get_device(
        cmd: commands.GetDevice,
        uow: unit_of_work.AbstractUnitOfWork,
):
    with uow:
        uow.do(cmd)


def get_state(
        cmd: commands.GetState,
        uow: unit_of_work.AbstractUnitOfWork,
):
    with uow:
        uow.do(cmd)


def publish_state_retrieved(
        event: events.StateRetrieved,
        publish: Callable
):
    publish(event)


def publish_device_retrieved(
        event: events.DeviceRetrieved,
        publish: Callable
):
    publish(event)


EVENT_HANDLERS = {
    events.AOpened: [
        publish_a_opened,
    ],
    events.BOpened: [
        publish_b_opened,
    ],
    events.FlowedA: [
        publish_flow_a,
    ],
    events.FlowedB: [
        publish_flow_b,
    ],
    events.StateRetrieved: [
        publish_state_retrieved,
    ],
    events.DeviceRetrieved: [
        publish_device_retrieved,
    ],
}

COMMAND_HANDLERS = {
    commands.OpenA: open_a,
    commands.OpenB: open_b,
    commands.Flow: flow,
    commands.GetDevice: get_device,
    commands.GetState: get_state,
}
