import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { deleteTransition, getTransition } from "~/models/transition.server";
import { Form, NavLink, Outlet, useLoaderData } from "@remix-run/react";

export const action = async ({ params }: LoaderArgs) => {
  invariant(params.transitionID, "transitionID not found");
  await deleteTransition({id: params.transitionID});
  return redirect(`/design/${params.netID}/transitions`);
}

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.transitionID, "netID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const transition = await getTransition({ id: params.transitionID });
  return json({ transition: transition });
};

export default function Transition() {
  const { transition } = useLoaderData<typeof loader>();

  const creation = new Date(transition.createdAt).toLocaleString();
  const updated = new Date(transition.updatedAt).toLocaleString();

  return (
    <div className={"flex flex-col justify-center space-y-2 p-2"}>
      <h1 className={"text-2xl font-bold"}>{transition.name}</h1>
      <p className={"text-xs justify-start"}>ID: {transition.id}</p>
      <p className={"text-sm"}>{transition.description}</p>
      <p className={"text-sm justify-start"}>Created: {creation}</p>
      <p className={"text-sm"}>Updated: {updated}</p>
      <br />
      <div className={"flex flex-row flex-shrink justify-between"}>
        <NavLink to={"edit"} className={"rounded-full contents-center bg-slate-600 text-white justify-center flex px-4 py-2"}>
          Edit
        </NavLink>
        <Form method={"post"}>
          <button type={"submit"} className={"rounded-full contents-center bg-red-800 text-white justify-center flex px-4 py-2"}>
          Delete
          </button>
        </Form>
      </div>

      <div >
        <Outlet />
      </div>
    </div>
  );
}