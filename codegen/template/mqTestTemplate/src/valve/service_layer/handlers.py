from valve.domain import commands
from valve.service_layer import unit_of_work


def open_a(
        cmd: commands.OpenA,
        uow: unit_of_work.AbstractUnitOfWork,
):
    with uow:
        uow.do(cmd)


EVENT_HANDLERS = {

}

COMMAND_HANDLERS = {
    commands.OpenA: open_a,
}
