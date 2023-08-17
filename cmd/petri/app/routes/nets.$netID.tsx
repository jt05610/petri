import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { isRouteErrorResponse, NavLink, Outlet, useLoaderData, useRouteError } from "@remix-run/react";
import invariant from "tiny-invariant";
import { getNet } from "~/models/net.server";
import { requireUserId } from "~/session.server";
import { useMemo, useState } from "react";
import { PetriNet } from "~/util/petrinet";
import { LabeledNet } from "~/lib/components/labeledNet";


export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.netID, "netID not found");

  const net = await getNet({ id: params.netID, authorID: authorID });
  if (!net) {
    throw new Response("Not Found", { status: 404 });
  }

  return json({ net: net });
};

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
  const petriNet = useMemo(() => {
    return new PetriNet(data.net);
  }, [data.net]);
  return (
    <div className={"flex flex-col w-full h-full space-y-2"}>
      <div className="flex flex-col w-full h-5/6 items-center">
        <LabeledNet net={petriNet} />
      </div>
      <div className={"flex flex-row w-full h-1/4 bg-slate-100 dark:bg-slate-700 space-y-1 p-2"}>
        <div className={"flex flex-col w-1/3 h-full space-y-1"}>
          <h3 className="text-lg font-bold">{data.net.name}</h3>
          <p className="py-6 text-md">{data.net.description}</p>
        </div>
        <div className={"flex flex-col w-2/3 space-y-1"}>
          <div>
            <LinkList btnClass={"text-white p-2"} routes={[
              { name: "Places", path: "places" },
              { name: "Transitions", path: "transitions" },
              { name: "Arcs", path: "arcs" },
              { name: "Sequence", path: `/control/${data.net.id}/record` },
            ]} />
          </div>
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
