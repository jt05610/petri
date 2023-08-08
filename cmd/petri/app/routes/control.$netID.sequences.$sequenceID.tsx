import type { ReactNode } from "react";
import { PencilIcon, PlayIcon } from "@heroicons/react/24/solid";
import { NavLink, Outlet, useLoaderData } from "@remix-run/react";
import { RunProvider } from "~/lib/context/run";
import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import invariant from "tiny-invariant";
import { requireUser } from "~/session.server";
import { getRunDetails } from "~/models/net.run.server";
import { Suspense } from "react";

type NavLinkBoxItem = {
  to: string;
  label: string;
  icon: ReactNode;
}

export function NavLinkBoxItem(props: NavLinkBoxItem) {
  return (
    <div className="flex flex-col items-center justify-center">
      <NavLink to={props.to}>
        {props.icon}
        <span className="text-sm">
        {props.label}
      </span>
      </NavLink>
    </div>
  );
}

type NavLinkBoxProps = {
  items: NavLinkBoxItem[];
}

export function NavLinkBox(props: NavLinkBoxProps) {
  return (
    <div className="flex flex-row justify-around">
      {props.items.map((item, i) => <NavLinkBoxItem key={i} {...item} />)}
    </div>
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
    <div className="flex flex-row items-center justify-start">
      <div className="flex flex-row items-center">
        <h1 className="text-2xl font-bold">Sequence</h1>
        <NavLinkBox items={[
          {
            to: "edit",
            label: "Edit",
            icon: <PencilIcon className="h-6 w-6" />
          },
          {
            to: "play",
            label: "Play",
            icon: <PlayIcon className="h-6 w-6" />
          }
        ]} />
      </div>
      <Suspense fallback={<div>Loading...</div>}>
        {run &&
          <RunProvider runDetails={run}>
            <Outlet />
          </RunProvider>
        }
      </Suspense>
    </div>
  );
}
