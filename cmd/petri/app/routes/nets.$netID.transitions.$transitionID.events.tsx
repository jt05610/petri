import { requireUserId } from "~/session.server";
import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { listEvents } from "~/models/net.transition.event.server";
import { NavLink, Outlet, useLoaderData } from "@remix-run/react";


export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.transitionID, "transitionID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const events = await listEvents({ transitionID: params.transitionID });
  return json({ events: events });
};

export default function Events() {
  const data = useLoaderData<typeof loader>();
  return (
    <div>
      <div className={"flex flex-col items-center justify-start space-y-1 p-2"}>
        <h3 className="text-2xl font-bold">Events</h3>
        {data.events.map((event) => (
          <div className={`flex flex-row justify-start w-full`} key={event.id}>
            <NavLink to={event.id}
                     key={event.id}>
              {event.name}
            </NavLink>
          </div>
        ))}
      </div>
      <Outlet />
    </div>
  );
};