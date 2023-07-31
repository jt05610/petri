import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { isRouteErrorResponse, NavLink, Outlet, useLoaderData, useRouteError } from "@remix-run/react";
import invariant from "tiny-invariant";
import type { FormInput } from "~/lib/components/formInput";
import { getNet } from "~/models/net.server";
import { requireUserId } from "~/session.server";
import { MermaidDiagram } from "~/lib/components/mermaid";
import { NetProvider } from "~/context/net";
import { useEffect, useState } from "react";

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.netID, "netID not found");

  const net = await getNet({ id: params.netID, authorID: authorID });
  if (!net) {
    throw new Response("Not Found", { status: 404 });
  }

  return json({ net: net });
};

export default function NetDetailsPage() {
  const data = useLoaderData<typeof loader>();
  const [net, setNet] = useState(data.net);
  const [selected, setSelected] = useState({ kind: "place", id: data.net.places[0].id });
  const [formData, setFormData] = useState<FormInput>({ fromPlace: false, id: "", name: "", description: "" });
  useEffect(() => {
    setNet(data.net);
  }, [data.net]);

  return (
    <NetProvider value={data.net}>
      <div className={"flex flex-row w-full h-full space-y-2"}>
        <div className={"flex flex-col w-1/4 bg-slate-100 p-2"}>
          <h3 className="text-2xl font-bold">{data.net.name}</h3>
          <p className="py-6">{data.net.description}</p>
          <hr />
          <h3 className="text-2xl justify-start font-bold">Actions</h3>
          <h4 className="text-xl font-bold">Create</h4>
          <div className={"flex flex-row w-full items-center justify-center space-x-1"}>
            <NavLink
              to={"add-place"}
              className={"rounded bg-slate-600 text-white p-2"}
            >
              <div>
                Place
              </div>
            </NavLink>
            <NavLink
              className={"rounded bg-slate-600 text-white p-2"}
              to={"add-transition"}>
              Transition
            </NavLink>
            <NavLink
              className={"rounded bg-slate-600 text-white p-2"}
              to={"add-arc"}>Arc</NavLink>
          </div>
          <Outlet context={[selected, formData, setFormData]} />
        </div>
        <div className="flex flex-col w-full h-full items-center">
          <MermaidDiagram net={net} setSelected={setSelected} />
        </div>
      </div>
    </NetProvider>
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
