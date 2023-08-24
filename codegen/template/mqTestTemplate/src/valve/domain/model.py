from enum import Enum
from typing import List, Dict

from . import events


class ValvePosition(str, Enum):
    A = "A"
    B = "B"

    def __str__(self) -> str:
        return str.__str__(self)


class Valve:
    events: List[events.Event]
    position: ValvePosition
    state: Dict[str, int]

    def __init__(self, initial_position: ValvePosition):
        self.events = []
        self.position = initial_position
        self.state = {}

    def open_a(self):
        self.position = ValvePosition.A
        self.events.append(events.AOpened())

    def open_b(self):
        self.position = ValvePosition.B
        self.events.append(events.BOpened())

    def flow(self, volume: float):
        if self.position == ValvePosition.A:
            self.events.append(events.FlowedA(volume))
        elif self.position == ValvePosition.B:
            self.events.append(events.FlowedB(volume))
        else:
            raise Exception("Valve position is not A or B")

    def get_device(self):
        self.events.append(events.DeviceRetrieved())

    def get_state(self):
        self.events.append(events.StateRetrieved(self.state))
