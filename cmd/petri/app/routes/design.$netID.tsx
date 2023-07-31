import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { Tab } from "@headlessui/react";
import { isRouteErrorResponse, NavLink, Outlet, useLoaderData, useRouteError, useSubmit } from "@remix-run/react";
import invariant from "tiny-invariant";
import { getNet } from "~/models/net.server";
import { requireUserId } from "~/session.server";
import { MermaidDiagram } from "~/lib/components/mermaid";
import { FormProvider} from "~/context/form";

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

type NodeListProps = {
  nodes: {
    "id": string,
    "name": string,
  }[]
  kind: "place" | "transition"
}

function NodeList(props: NodeListProps) {
  const submit = useSubmit();

  function handleDelete(id: string) {
    submit({ id: id }, { method: "post", action: `delete-${props.kind}`, encType: "application/json" });
  }

  return (
    <div className={"flex flex-col w-full items-center justify-center space-y-2 p-2"}>
      {props.nodes.map((node) => (
        <div className={`flex flex-row justify-between w-full `} key={node.id}>
          <NavLink to={`update-${props.kind}?id=${node.id}`} className={"p-2"}
                   key={node.id}>
            {node.name}
          </NavLink>
          <button className={"rounded bg-red-800 text-white p-1"} onClick={() => handleDelete(node.id)}>
            X
          </button>
        </div>
      ))}
    </div>
  );
}

export default function NetDetailsPage() {
  const data = useLoaderData<typeof loader>();
  const clicked = (id: string, kind: "place" | "transition") => {

  };

  return (
    <FormProvider>
      <div className={"flex flex-row w-full h-full space-y-2"}>
        <div className={"flex flex-col w-1/4 bg-slate-100 p-2"}>
          <h3 className="text-2xl font-bold">{data.net.name}</h3>
          <p className="py-6">{data.net.description}</p>
          <hr />
          <h3 className="text-2xl justify-start font-bold">Actions</h3>
          <h4 className="text-xl font-bold">Create</h4>
          <LinkList btnClass={"rounded bg-slate-600 text-white p-2"} routes={[
            { name: "Place", path: "add-place" },
            { name: "Transition", path: "add-transition" },
            { name: "Arc", path: "add-arc" }
          ]} />
          <hr />
          <Tab.Group>
            <Tab.List className={"flex flex-row items center"}>
              <h4 className="text-xl px-4 py-2 font-bold">View</h4>
              <Tab className={"text-lg px-4 py-2"}>Places</Tab>
              <Tab className={"text-lg px-4 py-2"}>Transitions</Tab>
              <Tab className={"text-lg px-4 py-2"}>Events</Tab>
            </Tab.List>
            <Tab.Panels>
              <Tab.Panel>
                <NodeList nodes={data.net.places} kind={"place"} />
              </Tab.Panel>
              <Tab.Panel>
                <NodeList nodes={data.net.transitions} kind={"transition"} />
              </Tab.Panel>
              <Tab.Panel>
                <NodeList nodes={data.net.transitions} kind={"transition"} />
              </Tab.Panel>
            </Tab.Panels>
          </Tab.Group>
          <Outlet />
        </div>
        <div className="flex flex-col w-full h-full items-center">
          <MermaidDiagram net={data.net} clicked={clicked} />
        </div>
      </div>
    </FormProvider>
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
