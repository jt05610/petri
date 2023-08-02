import { requireUserId } from "~/session.server";
import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { listPlaces } from "~/models/place.server";
import { NavLink, Outlet, useLoaderData } from "@remix-run/react";


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
    <div>
      <div className={"flex flex-col items-center justify-start space-y-1 p-2"}>
        <h3 className="text-2xl font-bold">Places</h3>
        {data.places.map((place) => (
          <div className={`flex flex-row justify-start w-full`} key={place.id}>
            <NavLink to={place.id}
                     key={place.id}>
              {place.name}
            </NavLink>
          </div>
        ))}
      </div>
      <Outlet />
    </div>
  );
};