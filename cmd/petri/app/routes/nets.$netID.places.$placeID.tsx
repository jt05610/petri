import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { deletePlace, getPlace } from "~/models/net.place.server";
import { Form, NavLink, Outlet, useLoaderData } from "@remix-run/react";

export const action = async ({ params }: LoaderArgs) => {
  invariant(params.placeID, "placeID not found");
  await deletePlace({id: params.placeID});
  return redirect(`/nets/${params.netID}/places`);
}

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.placeID, "netID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const place = await getPlace({ id: params.placeID });
  return json({ place: place });
};

export default function Place() {
  const { place } = useLoaderData<typeof loader>();

  const creation = new Date(place.createdAt).toLocaleString();
  const updated = new Date(place.updatedAt).toLocaleString();

  return (
    <div className={"flex flex-col justify-center space-y-2 p-2"}>
      <h1 className={"text-2xl font-bold"}>{place.name}</h1>
      <p className={"text-xs justify-start"}>ID: {place.id}</p>
      <p className={"text-sm"}>{place.description}</p>
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