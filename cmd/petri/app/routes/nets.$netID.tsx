import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { isRouteErrorResponse, NavLink, Outlet, useLoaderData, useRouteError } from "@remix-run/react";
import invariant from "tiny-invariant";
import { getNet } from "~/models/net.server";
import { requireUserId } from "~/session.server";

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
    <div className={"flex flex-row w-full items-center justify-center space-x-1"}>
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

  return (
    <div className={"flex flex-row w-full h-full space-y-2"}>
      <div className={"flex flex-col w-1/4 bg-slate-100 space-y-1 p-2"}>
        <h3 className="text-2xl font-bold">{data.net.name}</h3>
        <p className="py-6">{data.net.description}</p>
        <h3 className="text-2xl font-bold">Subsystems</h3>
        <hr className="rounded-full border-2 border-slate-200" />
        <div>
          {data.net.children.map((child, i) => (
            <div key={child.id}>
              <NavLink to={`/design/${child.id}`}>{i + 1}. {child.name}</NavLink>
            </div>
          ))}
        </div>
        <h3 className="text-2xl font-bold">Actions</h3>
        <hr className="rounded-full border-2 border-slate-200" />
        <div>
          <LinkList btnClass={"rounded bg-slate-600 text-white p-2"} routes={[
            { name: "Places", path: "places" },
            { name: "Transitions", path: "transitions" },
            { name: "Arcs", path: "arcs" },
          ]} />
          <Outlet />
        </div>
      </div>
      <div className="flex flex-col w-full h-full items-center">
      </div>
    </div>
  )
    ;
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
