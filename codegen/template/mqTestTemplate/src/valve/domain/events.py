from dataclasses import dataclass


@dataclass
class Event:
    pass


@dataclass
class AOpened(Event):
    def __str__(self):
        return "open_a"


@dataclass
class BOpened(Event):
    def __str__(self):
        return "open_b"


@dataclass
class FlowedA(Event):
    volume: float

    def __str__(self):
        return "flowed_a"


@dataclass
class FlowedB(Event):
    volume: float

    def __str__(self):
        return "flowed_b"
