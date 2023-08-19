import { NavLink } from "@remix-run/react";
import { useContextSelector } from "use-context-selector";
import { ControlContext } from "~/routes/control";

type NetCardProps = {
  net: {
    id: string;
    name: string;
    description: string;
  };
}

export function NetCard({ net }: NetCardProps) {
  return (
    <div
      className={"flex flex-col border w-1/4 space-y-2 border-gray-300 rounded-lg p-2 my-4 shadow-md hover:shadow-lg transition-shadow duration-300 ease-in-out h-100"}
    >
      <h4
        className={"text-xl font-semibold"}
      >
        {net.name}
      </h4>
      <hr className={"rounded-full my-2"} />
      <p className={"overflow-auto"}>{net.description}</p>
      <div className={"flex flex-row space-x-2 w-full justify-center"}>
        <NavLink
          to={`${net.id}/record`}
          className={"rounded-full w-min bg-slate-600 px-4 py-2 text-blue-100 hover:bg-rose-500 active:bg-rose-600 mt-auto"}
        >
          Record
        </NavLink>
        <NavLink
          to={`${net.id}/sequences`}
          className={"rounded-full w-min bg-slate-600 px-4 py-2 text-blue-100 hover:bg-teal-500 active:bg-teal-600 mt-auto"}
        >
          Sequences
        </NavLink>
      </div>
    </div>
  );
}

type NetGroupProps = {
  title: string;
  nets: {
    id: string;
    name: string;
    description: string;
  }[]
}

export function NetGroup({ title, nets }: NetGroupProps) {
  return (
    <div className={"flex flex-col p-2 w-full"}>
      <h3
        className={"text-2xl font-semibold"}
      >
        {title}
      </h3>
      <div className={"flex flex-wrap space-x-2 w-full h-full"}>
        {nets.map((net) => {
            return (
              <NetCard net={net} key={net.id} />
            );
          }
        )}
      </div>
    </div>
  );
}

export default function ControlIndex() {
  const { systemNets, deviceNets } = useContextSelector(ControlContext, (ctx) => ctx!);
  return (
    <div className={"flex flex-col p-2 w-full h-full"}>
      <div className={"p-2 space-y-2"}>
        <h2
          className={"text-4xl font-semibold"}
        >Record
        </h2>
        <p>
          Record a sequence of events to be played back later. Select either a system or a device to record.
        </p>
      </div>

      <div
        className={"flex h-1/2 w-full"}
      >
        <NetGroup
          title={"Systems"}
          nets={systemNets}
        />
      </div>
      <div
        className={"flex h-1/2 w-full"}
      >
        <NetGroup
          title={"Devices"}
          nets={deviceNets}
        />
      </div>
    </div>
  );
}