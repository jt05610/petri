import asyncio
import logging
from typing import Optional

import grpc
from autosampler.entrypoints import autosampler_pb2, autosampler_pb2_grpc

from autosampler import comm, config

logger = logging.getLogger(__name__)


def convert_status(status: comm.AutosamplerStatus) -> autosampler_pb2.InjectState:
    if status.unk_2 == 3:
        if status.injected:
            return autosampler_pb2.InjectState.Injected
        return autosampler_pb2.InjectState.Injecting
    return autosampler_pb2.InjectState.Idle


class Autosampler(autosampler_pb2_grpc.AutosamplerServicer):
    last: Optional[comm.InjectionParameters]

    def __init__(self):
        self._ser = comm.SerialInterface(config.PORT, config.BAUD, config.TIMEOUT)
        self._ser.initialize_port()
        self.last = None

    async def Inject(
            self,
            request: autosampler_pb2.InjectRequest,
            context: grpc.aio.ServicerContext
    ) -> autosampler_pb2.InjectResponse:
        logger.info("Injecting %s", request)
        params = comm.InjectionParameters(
            request.vial,
            request.air_cushion,
            request.excess_volume,
            request.flush_volume,
            request.injection_volume,
            request.needle_depth,
        )
        self._ser.reset_status()
        commands = params.inject_commands(self.last)
        self._ser.send_commands(commands)
        self.last = params
        status = self._ser.poll()
        while status.unk_2 == 3:
            status = self._ser.poll()
            if not isinstance(status.unk_2, int):
                status.unk_2 = 3
                continue
            yield autosampler_pb2.InjectResponse(
                state=convert_status(status),
            )


async def serve() -> None:
    server = grpc.aio.server()
    autosampler_pb2_grpc.add_AutosamplerServicer_to_server(Autosampler(), server)
    listen = server.add_insecure_port('[::]:50051')
    await server.start()
    logger.info("Listening on %s", listen)
    await server.wait_for_termination()


if __name__ == '__main__':
    logging.basicConfig(level=logging.INFO)
    asyncio.run(serve())
