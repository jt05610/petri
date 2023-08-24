from __future__ import annotations
import abc
import threading
from typing import Callable, Dict, Type

import serial

from valve import config
from valve.domain import commands
from valve.domain.model import Valve


class AbstractUnitOfWork(abc.ABC):
    valve: Valve
    mutex = threading.Lock()
    handlers: Dict[Type[commands.Command], Callable] = {
        commands.OpenA: Valve.open_a,
        commands.OpenB: Valve.open_b,
        commands.Flow: Valve.flow,
        commands.GetDevice: Valve.get_device,
        commands.GetState: Valve.get_state,
    }

    def __enter__(self) -> AbstractUnitOfWork:
        self.mutex.acquire()
        return self

    def __exit__(self, *args):
        self.mutex.release()
        return

    def collect_new_events(self):
        while self.valve.events:
            yield self.valve.events.pop(0)

    def _do(self, cmd: commands.Command):
        raise NotImplementedError

    def do(self, cmd: commands.Command):
        if type(cmd) in [commands.OpenA, commands.OpenB]:
            self._do(cmd)

        self.handlers[type(cmd)](self.valve)


class GRBLUnitOfWork(AbstractUnitOfWork):
    MSG = {
        commands.OpenA: config.OPEN_A_COMMAND,
        commands.OpenB: config.OPEN_B_COMMAND,
    }
    port: serial.Serial

    def __init__(self, valve: Valve, ser: str,
                 baud: int = 115200,
                 timeout: int = 1):
        self.valve = valve
        self.port = serial.serial_for_url(ser)

    def write(self, cmd: commands.Command):
        msg = self.MSG[type(cmd)] + "\n"
        self.port.write(msg.encode('utf-8'))

    def read(self):
        return self.port.readline().decode('utf-8').strip()

    def _do(self, cmd: commands.Command):
        self.write(cmd)
        ret = self.read()
        if ret != "ok":
            raise Exception(f"GRBL returned {ret}")
