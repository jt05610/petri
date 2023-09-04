import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { deleteArc, getArc } from "~/models/net.arc.server";
import { Form, NavLink, Outlet, useLoaderData } from "@remix-run/react";

export const action = async ({ params }: LoaderArgs) => {
  invariant(params.arcID, "arcID not found");
  await deleteArc({ id: params.arcID });
  return redirect(`/nets/${params.netID}/arcs`);
};

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.arcID, "netID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const arc = await getArc({ id: params.arcID });
  return json({ arc: arc });
};

export default function Arc() {
  const { arc } = useLoaderData<typeof loader>();

  const creation = new Date(arc.createdAt).toLocaleString();
  const updated = new Date(arc.updatedAt).toLocaleString();

  return (
    <div className={"flex flex-col justify-center space-y-2 p-2"}>
      <h1
        className={"text-2xl font-bold"}>{arc.fromPlace ? arc.place.name + " -> " + arc.transition.name : arc.transition.name + " -> " + arc.place.name}</h1>
      <p className={"text-xs justify-start"}>ID: {arc.id}</p>
      <p className={"text-sm justify-start"}>Created: {creation}</p>
      <p className={"text-sm"}>Updated: {updated}</p>
      <br />
      <div className={"flex flex-row flex-shrink justify-between"}>
        <NavLink to={"edit"}
                 className={"rounded-full contents-center bg-slate-600 text-white justify-center flex px-4 py-2"}>
          Edit
        </NavLink>
        <Form method={"post"}>
          <button type={"submit"}
                  className={"rounded-full contents-center bg-red-800 text-white justify-center flex px-4 py-2"}>
            Delete
          </button>
        </Form>
      </div>

      <div>
        <Outlet />
      </div>
    </div>
  );
}