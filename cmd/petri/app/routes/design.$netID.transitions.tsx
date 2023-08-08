import { requireUserId } from "~/session.server";
import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { listTransitions } from "~/models/net.transition.server";
import { NavLink, Outlet, useLoaderData } from "@remix-run/react";


export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.netID, "netID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const transitions = await listTransitions({ netID: params.netID });
  return json({ transitions: transitions });
};

export default function Transitions() {
  const data = useLoaderData<typeof loader>();
  return (
    <div>
      <div className={"flex flex-col items-center justify-start space-y-1 p-2"}>
        <h3 className="text-2xl font-bold">Transitions</h3>
        {data.transitions.map((transition) => (
          <div className={`flex flex-row justify-start w-full`} key={transition.id}>
            <NavLink to={transition.id}
                     key={transition.id}>
              {transition.name}
            </NavLink>
          </div>
        ))}
      </div>
      <Outlet />
    </div>
  );
};