import { PencilIcon, PlayIcon } from "@heroicons/react/24/solid";
import { NavLink, Outlet, useLoaderData } from "@remix-run/react";
import { RunProvider } from "~/lib/context/run";
import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import invariant from "tiny-invariant";
import { requireUser } from "~/session.server";
import { getRunDetails } from "~/models/net.run.server";
import type { ReactNode } from "react";
import { Suspense } from "react";
import Dropdown, { DropdownItem } from "~/lib/components/dropdown";


type ActionDropdownItemProps = {
  active: boolean
  key: string
  children: ReactNode
  route: string
}

function ActionDropdownItem({ route, active, key, children }: ActionDropdownItemProps) {
  return (
    <NavLink to={route}>
      <DropdownItem key={key} isActive={active}>
        <div
          className={"group flex w-full items-center rounded-md px-2 py-2 text-sm"}
        >
          {children}
        </div>
      </DropdownItem>
    </NavLink>
  );
}

function ActionDropdown() {
  return (
    <Dropdown title={"Actions"}>
      <ActionDropdownItem key={"edit"} active={false} route={"edit"}>
        <PencilIcon className="h-6 w-6" />
        Edit
      </ActionDropdownItem>
      <ActionDropdownItem active={false} key={"play"} route={"play"}>
        <PlayIcon className="h-6 w-6" />
        Play
      </ActionDropdownItem>

    </Dropdown>
  );
}


export const loader = async ({ params, request }: LoaderArgs) => {
  const { sequenceID } = params;
  invariant(sequenceID, "sequenceID is required");
  await requireUser(request);
  const details = await getRunDetails({ runID: sequenceID });
  return json({ run: details });
};

export default function SequencePage() {
  const { run } = useLoaderData<typeof loader>();

  return (
    <div className="p-2 flex justify-start flex-col w-full h-full space-y-2">
      <div className="flex items-center flex-row justify-between">
        <h1 className="text-3xl font-bold">{run.name}</h1>
        <ActionDropdown />
      </div>
      <div className="flex flex-row space-x-2">
        <Suspense fallback={<div>Loading...</div>}>
          {run &&
            <RunProvider runDetails={run}>
              <Outlet />
            </RunProvider>
          }
        </Suspense>
      </div>
    </div>
  );
}
