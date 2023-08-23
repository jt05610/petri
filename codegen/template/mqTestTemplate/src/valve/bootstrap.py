import inspect
from typing import Callable

from valve import config
from valve.adapters import rabbitmq_publisher
from valve.domain import model
from valve.service_layer import handlers, messagebus, unit_of_work


def bootstrap(
        uow: unit_of_work.AbstractUnitOfWork = unit_of_work.GRBLUnitOfWork(
            valve=model.Valve(initial_position=config.INITIAL_POSITION),
            ser=config.SERIAL_PORT,
            baud=config.SERIAL_BAUD,
            timeout=config.SERIAL_TIMEOUT,
        ),
        publish: Callable = rabbitmq_publisher.publish,

) -> messagebus.MessageBus:
    dependencies = {"uow": uow,
                    "publish": publish}
    injected_event_handlers = {
        event_type: [
            inject_dependencies(handler, dependencies)
            for handler in event_handlers
        ]
        for event_type, event_handlers in handlers.EVENT_HANDLERS.items()
    }
    injected_command_handlers = {
        command_type: inject_dependencies(handler, dependencies)
        for command_type, handler in handlers.COMMAND_HANDLERS.items()
    }

    return messagebus.MessageBus(
        uow=uow,
        event_handlers=injected_event_handlers,
        command_handlers=injected_command_handlers,
    )


def inject_dependencies(handler, dependencies):
    params = inspect.signature(handler).parameters
    deps = {
        name: dependency
        for name, dependency in dependencies.items()
        if name in params
    }
    return lambda message: handler(message, **deps)
