import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { NavLink, Outlet, useLoaderData } from "@remix-run/react";
import type { RunListItem } from "~/models/net.run.server";
import { listRuns } from "~/models/net.run.server";
import { requireUserId } from "~/session.server";
import { getUserById } from "~/models/user.server";
import invariant from "tiny-invariant";
import Dropdown, { DropdownItem } from "~/lib/components/dropdown";

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  invariant(params.netID, "netID not found");
  const netID = params.netID;
  const sequences = await listRuns({ netID });

  return json({ sequences, netID });
};

type RunDropdownProps = {
  runs: RunListItem[]
  netID: string
}

type RunDropdownItemProps = {
  active: boolean
  netID: string
  run: RunListItem
  key: string
}

function RunDropdownItem({ netID, active, key, run }: RunDropdownItemProps) {
  return (
    <NavLink to={`/control/${netID}/sequences/${run.id}`}>
      <DropdownItem key={key} isActive={active}>
        {run.name}
      </DropdownItem>
    </NavLink>
  );
}

function RunDropdown({ runs, netID }: RunDropdownProps) {
  return (
    <Dropdown title={"Sequences"}>
      {runs.map((run) => (
        <RunDropdownItem key={run.id} run={run} active={false} netID={netID} />
      ))}
    </Dropdown>
  );
}

export default function RunsPage() {
  const { sequences, netID } = useLoaderData<typeof loader>();
  return (
    <div className="flex flex-col min-h-full w-full">
      <div className={"relative top-0 right-0 text-right"}>
        <RunDropdown runs={sequences} netID={netID} />
      </div>
      <Outlet />
    </div>
  );
}
