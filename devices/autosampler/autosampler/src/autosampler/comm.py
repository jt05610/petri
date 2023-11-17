from dataclasses import dataclass
from typing import Tuple, Iterable, Optional

import serial

from autosampler import config


@dataclass
class Command:
    string: bytes
    variable: int = None

    def __bytes__(self):
        if self.variable is not None:
            result = self.string + b"," + bytes(str(self.variable), 'utf-8') + b"\r"
        else:
            result = self.string + b"\r"
        return result


class IncorrectResponse(Exception):
    pass


CONNECT = Command(b"2")
CLOSE = Command(b"3")

WAIT = Command(b"S,33")
IDLE = Command(b"S,29")

DEVICE_ID = Command(b"G,10")
TRAY_STATUS = Command(b"G,12")


@dataclass
class AutosamplerStatus:
    tray: int
    unk_1: int
    unk_2: int
    injected: bool = False


class InjectionParameters:
    _vial: int
    _air_cushion: int
    _excess_volume: int
    _flush_volume: int
    _injection_volume: int
    _needle_depth: int
    _syringe: int = 250

    def __init__(self, vial: int, air_cushion: int, excess_volume: int, flush_volume: int, injection_volume: int,
                 needle_depth: int, syringe: int = 250):
        self._vial = vial
        self._air_cushion = air_cushion
        self._excess_volume = excess_volume
        self._flush_volume = flush_volume
        self._injection_volume = injection_volume
        self._needle_depth = needle_depth if needle_depth <= 0 else -needle_depth
        self._syringe = syringe

    @property
    def vial(self) -> Command:
        return Command(b"S,30", self._vial)

    @vial.setter
    def vial(self, v: int):
        self._vial = v

    @property
    def needle_depth(self) -> Command:
        return Command(b"S,26", self._needle_depth)

    @needle_depth.setter
    def needle_depth(self, nd: int):
        self._needle_depth = nd if nd <= 0 else -nd

    @property
    def air_cushion(self) -> Command:
        return Command(b"S,21", self._air_cushion)

    @air_cushion.setter
    def air_cushion(self, ac: int):
        self._air_cushion = ac

    @property
    def excess_volume(self) -> Command:
        return Command(b"S,20", self._excess_volume)

    @excess_volume.setter
    def excess_volume(self, xv: int):
        self._excess_volume = xv

    @property
    def flush_volume(self) -> Command:
        return Command(b"S,22", self._flush_volume)

    @flush_volume.setter
    def flush_volume(self, fv: int):
        self._flush_volume = fv

    @property
    def injection_volume(self) -> Command:
        return Command(b"S,33", self._injection_volume)

    @injection_volume.setter
    def injection_volume(self, iv: int):
        self._injection_volume = iv

    def setup_gen(self, last: "InjectionParameters") -> Iterable[Command]:
        params = (
            "air_cushion",
            "excess_volume",
            "flush_volume",
            "needle_depth",
            "injection_volume",
        )
        updated = False
        for p in params:
            if getattr(self, p) != getattr(last, p):
                updated = True
                yield getattr(self, p)
        if updated:
            rest = (
                Command(b"S,23", 1),
                Command(b"S,28,0,0"),
                Command(b"S,24", 3),
                Command(b"S,90", 0),
                Command(b"S,91", 0),
                Command(b"S,34", 1),
                Command(b"S,35,0,0"),
                Command(b"S,29", 0),
                Command(b"S,27", 0),
            )
            for r in rest:
                yield r

    def inject_commands(self, last: Optional["InjectionParameters"] = None) -> Tuple[Command, ...]:
        if last is None:
            return self.commands
        setup = self.setup_gen(last)
        return tuple(setup) + (self.vial, Command(b"1"))

    @property
    def commands(self) -> Tuple[Command, ...]:
        return (self.vial,
                self.air_cushion,
                self.excess_volume,
                Command(b"S,23", 1),
                Command(b"S,28,0,0"),
                self.needle_depth,
                self.flush_volume,
                Command(b"S,24", 3),
                Command(b"S,90", 0),
                Command(b"S,91", 0),
                Command(b"S,34", 1),
                Command(b"S,35,0,0"),
                Command(b"S,29", 0),
                Command(b"S,27", 0),
                self.injection_volume,
                Command(b"1")
                )


class SerialInterface:
    _serial: serial.Serial
    _status: AutosamplerStatus

    def __init__(self, port: str, baud: int, timeout: float):
        self._serial = serial.Serial(port, baud, timeout=timeout)
        self._status = AutosamplerStatus(0, 0, 0)


    def reset_status(self):
        self._status = AutosamplerStatus(0, 0, 0)

    def _read(self) -> bytes:
        return self._serial.read(config.READ_SIZE)

    def _write(self, command: Command) -> bytes:
        self._serial.write(bytes(command))
        ret = self._read()
        print("Wrote: ", bytes(command), "Received: ", ret)
        return ret

    def _close(self):
        self._serial.close()

    def _write_expect_response(self, command: Command, response: bytes):
        resp = self._write(command)
        if response != resp.strip(b"\r"):
            return

    def _request_id(self):
        resp = b""
        read = self._write(DEVICE_ID).strip(b"\r")
        while len(read) > 0:
            resp += read
            read = self._read().strip(b"\r")
        return resp

    def _wait(self, time: int):
        command = Command(b"S,33", time)
        return self._write(command)

    def _set_idle(self, idle: bool):
        return self._write(Command(b"S,29", int(idle)))

    def check_tray(self):
        return self._write(Command(b"G,17"))

    def send_commands(self, commands: Tuple[Command, ...]):
        for c in commands:
            self._write_expect_response(c, b"0")

    @staticmethod
    def response(resp: bytes) -> Tuple[int, bool]:
        try:
            result = int(resp.strip(b'\r').split(b",")[-1])
        except ValueError:
            result = resp
        injected = b"I\n" in resp
        return result, injected

    def poll(self) -> AutosamplerStatus:
        self._status.tray, injected = self.response(self._write(TRAY_STATUS))
        if injected:
            self._status.injected = True
        self._status.unk_1, injected = self.response(self._write(Command(b"G,17")))

        if injected:
            self._status.injected = True
        self._status.unk_2, injected = self.response(self._write(Command(b"G,13")))

        if injected:
            self._status.injected = True
        return self._status

    def initialize_port(self):
        self._serial.flush()
        resp = self._request_id()
        while b"9" == resp:
            resp = self._request_id()

        self._write_expect_response(CONNECT, b"0")

        assert 0 == int(self._set_idle(True))

        for t in (2500, 1000, 500, 250, 100, 50):
            r = self._wait(t)

        assert 0 == int(r)

        assert b"G,17,0" == self.check_tray().strip(b"\r")

        self._write_expect_response(Command(b"E"), b"0")

        return resp

    def close_port(self):
        self._write_expect_response(CLOSE, b"0")
        self._close()
