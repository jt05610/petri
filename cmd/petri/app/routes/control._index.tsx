import { NavLink, useLoaderData } from "@remix-run/react";
import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import { getUserById } from "~/models/user.server";
import { getNetsWithDevice, getNetsWithEvents } from "~/models/net.server";

export const loader = async ({ request }: LoaderArgs) => {
  const userID = await requireUserId(request);
  const user = await getUserById(userID);
  if (!user) {
    throw new Error("User not found");
  }
  const systemNets = await getNetsWithEvents({ authorID: userID });
  const deviceNets = await getNetsWithDevice({ authorID: userID });
  console.log(systemNets);
  return json({ systemNets, deviceNets });
};

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
      <NavLink
        to={net.id}
        className={"rounded-full w-min bg-slate-600 px-4 py-2 text-blue-100 hover:bg-blue-500 active:bg-blue-600 mt-auto"}
      >
        Record
      </NavLink>
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

export default function RecordRoute() {
  const { systemNets, deviceNets } = useLoaderData<typeof loader>();
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