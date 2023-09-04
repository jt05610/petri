import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { isRouteErrorResponse, NavLink, Outlet, useLoaderData, useRouteError } from "@remix-run/react";
import invariant from "tiny-invariant";
import { getNet } from "~/models/net.server";
import { requireUserId } from "~/session.server";
import { useMemo, useState } from "react";
import { LabeledNet } from "~/lib/components/labeledNet";
import { Listbox } from "@headlessui/react";
import { useContextSelector } from "use-context-selector";
import { NetsContext } from "~/lib/context/nets";
import { PetriNet } from "~/util/petrinet";


export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.netID, "netID not found");

  const net = await getNet({ id: params.netID, authorID: authorID });
  if (!net) {
    throw new Response("Not Found", { status: 404 });
  }

  return json({ net: net });
};

type NetDropdownProps = {
  value: string[]
  setValue: (value: string[]) => void
  options?: string[]
}

function NetDropdown({ value, setValue, options }: NetDropdownProps) {
  return (
    <Listbox value={value} onChange={setValue} multiple>
      <Listbox.Button>
        {value.map((v) => (
          <span key={v}>{v}</span>
        ))}
      </Listbox.Button>
      <Listbox.Options>
        {options?.map((option) => (
          <Listbox.Option key={option} value={option}>
            {({ selected }) => (
              <div className={"flex flex-row w-full items-start justify-right text-right space-x-1"}>
                <span>{option}</span>
              </div>
            )}
          </Listbox.Option>
        ))}
      </Listbox.Options>
    </Listbox>
  );
}

type LinkListProps = {
  routes: {
    name: string;
    path: string;
  }[]
  btnClass: string;
}

function LinkList(props: LinkListProps) {
  return (
    <div className={"flex flex-row w-full items-start justify-right text-right space-x-1"}>
      {props.routes.map((route) => (
        <NavLink to={route.path} className={props.btnClass} key={route.name}>
          {route.name}
        </NavLink>
      ))}
    </div>
  );
}

export default function NetDetailsPage() {
  const data = useLoaderData<typeof loader>();
  const nets = useContextSelector(NetsContext, (context) => context);
  const [insertNets, setInsertNets] = useState<string[]>([]);
  const petriNet = useMemo(() => {
    return new PetriNet(data.net);
  }, [data.net]);
  return (
    <div className={"flex flex-row space-y-2"}>
      <LabeledNet net={petriNet} />
      <div className={"flex flex-col bg-slate-100 dark:bg-slate-700 space-y-1 p-2"}>
        <div className={"flex flex-col "}>
          <h3 className="text-lg font-bold">{data.net.name}</h3>
          <p className="text-md">{data.net.description}</p>
        </div>
        <NetDropdown value={insertNets} setValue={setInsertNets} options={nets?.map((net) => {
          return net.id;
        })} />
        <div>
          <LinkList btnClass={"text-white p-2"} routes={[
            { name: "Places", path: "places" },
            { name: "Transitions", path: "transitions" },
            { name: "Arcs", path: "arcs" },
            { name: "Sequence", path: `/control/${data.net.id}/record` },
            { name: "Devices", path: "devices" }
          ]} />
          <div className={"overflow-auto h-full"}>
            <Outlet />
          </div>
        </div>
      </div>
    </div>
  );
};


export function ErrorBoundary() {
  const error = useRouteError();

  if (error instanceof Error) {
    return <div>An unexpected error occurred: {error.message}</div>;
  }

  if (!isRouteErrorResponse(error)) {
    return <h1>Unknown Error</h1>;
  }

  if (error.status === 404) {
    return <div>Note not found</div>;
  }

  return <div>An unexpected error occurred: {error.statusText}</div>;
}
