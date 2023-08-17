import { requireUserId } from "~/session.server";
import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { listPlaces } from "~/models/net.place.server";
import { NavLink, Outlet, useLoaderData } from "@remix-run/react";
import { PlusIcon } from "@heroicons/react/24/outline";


export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.netID, "netID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const places = await listPlaces({ netID: params.netID });
  return json({ places: places });
};

export default function Places() {
  const data = useLoaderData<typeof loader>();
  return (
    <div className={"flex w-full h-full flex-row"}>
      <div className={"flex flex-col w-1/2 items-center justify-start text-start space-y-1 p-2"}>
        <div className={"flex flex-row font-bold justify-start w-full hover:text-teal-500"}>
          <NavLink to={"new"}>
            New
          </NavLink>
        </div>
        {data.places.map((place) => (
          <div className={`flex flex-row justify-start w-full`} key={place.id}>
            <NavLink to={place.id}
                     key={place.id}>
              {place.name}
            </NavLink>
          </div>
        ))}
      </div>
      <div className={"flex flex-col w-1/2 items-center justify-start text-start space-y-1 p-2"}>
        <Outlet />
      </div>
    </div>
  );
};