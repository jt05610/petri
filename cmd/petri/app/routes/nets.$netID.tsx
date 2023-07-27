import type { ActionArgs, LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import {
  Form,
  isRouteErrorResponse,
  useLoaderData,
  useRouteError
} from "@remix-run/react";
import invariant from "tiny-invariant";

import { deleteNet, getNet } from "~/models/net.server";
import { requireUserId } from "~/session.server";
import { MermaidDiagram } from "~/lib/components/mermaid";

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.netID, "noteId not found");

  const net = await getNet({ id: params.netID, authorID: authorID });
  if (!net) {
    throw new Response("Not Found", { status: 404 });
  }
  return json({ net: net });
};

export const action = async ({ params, request }: ActionArgs) => {
  const userId = await requireUserId(request);
  invariant(params.netID, "noteId not found");

  await deleteNet({ id: params.netID, authorID: userId });

  return redirect("/nets");
};

const Places = () => {
  const data = useLoaderData<typeof loader>();
  return (
    <div>
      <h3 className="text-xl font-bold">Places</h3>
      <ul className="list-disc list-inside">
        {data.net.places.map((place) => (
          <li key={place.id}>{place.name}</li>
        ))}
      </ul>
      <Form method="post" action="/add-place" className={`p-2`}>
        <input name="netID" type="hidden" value={data.net.id} />
        <div className={`flex flex-row p-2 space-x-2`}>

          <label
            htmlFor={`name`}
            className={`block text-sm font-medium text-gray-700`}
          >
            Name
          </label>
          <div>
            <input
              type={`text`}
              name={`name`}
            />
          </div>
        </div>
        <div className={`flex flex-row p-2 space-x-2`}>
          <label
            htmlFor={`bound`}
            className={`block text-sm font-medium text-gray-700`}
          >
            Bound
          </label>
          <div>
            <input
              type={`number`}
              name={`bound`}
              placeholder={`1`}
            />
          </div>
        </div>
        <button
          type="submit"
          className="rounded bg-blue-500 px-4 py-2 text-white hover:bg-blue-600 focus:bg-blue-400"
        >
          Add Place
        </button>
      </Form>

    </div>
  );
};

const initialCode = `
`;


export default function NetDetailsPage() {
  const data = useLoaderData<typeof loader>();

  const code = (): string => {
    if (data.net.places.length === 0) {
      return initialCode;
    }

    const places = data.net.places.map((place) => {
      return `${place.id}((${place.name}))\n  click ${place.id} clicked "${place.name}"`;
    });
    const transitions = data.net.transitions.map((transition) => {
      return `${transition.id}[${transition.name}]`;
    });
    const arcs = data.net.arcs.map((arc) => {
      console.log(arc);
      if (arc.fromPlace) {
        return `${arc.placeID} --> ${arc.transitionID}`;
      } else {
        return `${arc.transitionID} --> ${arc.placeID}`;
      }
    });
    let out = `graph TD
  ${places.join("\n  ")}
  ${transitions.join("\n  ")}
  ${arcs.join("\n  ")}
  `;
    console.log(out);
    return out;
  };
  //@ts-ignore
  window.clicked = async function(e:Event) {
    console.log(e);
  }

  return (
    <div>
      <h3 className="text-2xl font-bold">{data.net.name}</h3>
      <p className="py-6">{data.net.description}</p>
      <hr className="my-4" />
      <MermaidDiagram code={code()} />
      <Places />
      <Form method="post">
        <button
          type="submit"
          className="rounded bg-blue-500 px-4 py-2 text-white hover:bg-blue-600 focus:bg-blue-400"
        >
          Delete
        </button>
      </Form>
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
