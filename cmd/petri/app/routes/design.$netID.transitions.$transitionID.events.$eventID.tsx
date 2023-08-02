import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { deleteEvent, getEvent } from "~/models/event.server";
import { Form, NavLink, Outlet, useLoaderData } from "@remix-run/react";

export const action = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }

  invariant(params.eventID, "eventID not found");
  invariant(params.transitionID, "transitionID not found");
  invariant(params.netID, "transitionID not found");
  await deleteEvent({id: params.eventID});
  return redirect(`/design/${params.netID}/transitions/${params.transitionID}/events`);
}

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.eventID, "eventID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const event = await getEvent({ id: params.eventID });
  return json({ event: event });
};

export default function Event() {
  const { event } = useLoaderData<typeof loader>();

  const creation = new Date(event.createdAt).toLocaleString();
  const updated = new Date(event.updatedAt).toLocaleString();

  return (
    <div className={"flex flex-col justify-center space-y-2 p-2"}>
      <h1 className={"text-2xl font-bold"}>{event.name}</h1>
      <p className={"text-xs justify-start"}>ID: {event.id}</p>
      <p className={"text-sm"}>{event.description}</p>
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