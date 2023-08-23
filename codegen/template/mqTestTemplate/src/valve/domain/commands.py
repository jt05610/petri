from dataclasses import dataclass

from valve import config


class Command:
    pass


@dataclass
class OpenA(Command):
    def __str__(self):
        return "open_a"


@dataclass
class OpenB(Command):
    def __str__(self):
        return "open_b"


@dataclass
class Flow(Command):
    volume: float

    def __str__(self):
        return "flow"


@dataclass
class GetDevice(Command):
    def __str__(self):
        return config.DEVICE_ID


@dataclass
class GetState(Command):
    def __str__(self):
        return "state"
