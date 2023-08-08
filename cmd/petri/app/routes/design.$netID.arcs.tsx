import { requireUserId } from "~/session.server";
import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { listArcs } from "~/models/net.arc.server";
import { NavLink, Outlet, useLoaderData } from "@remix-run/react";


export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.netID, "netID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const arcs = await listArcs({ netID: params.netID });
  return json({ arcs: arcs });
};

export default function Arcs() {
  const data = useLoaderData<typeof loader>();
  return (
    <div>
      <div className={"flex flex-col items-center justify-start space-y-1 p-2"}>
        <h3 className="text-2xl font-bold">Arcs</h3>
        {data.arcs.map((arc) => (
          <div className={`flex flex-row justify-start w-full`} key={arc.id}>
            <NavLink to={arc.id}
                     key={arc.id}>
              {arc.fromPlace ? arc.place.name : arc.transition.name} {`-->`} {arc.fromPlace ? arc.transition.name : arc.place.name}
            </NavLink>
          </div>
        ))}
      </div>
      <Outlet />
    </div>
  );
};