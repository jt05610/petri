import type { LoaderArgs, ActionArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getNet } from "~/models/net.server";
import { Outlet, useLoaderData } from "@remix-run/react";
import { getUserById } from "~/models/user.server";
import { PetriNetProvider } from "~/lib/context/petrinet";

export const action = async ({ params, request }: ActionArgs) => {
  const userID = await requireUserId(request);
  invariant(params.netID, "netID not found");
  const user = await getUserById(userID);
  invariant(user, "User not found");
  console.log("user", user);
};

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.netID, "netID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const net = await getNet({ id: params.netID, authorID: authorID });
  if (!net) {
    throw new Response("Not Found", { status: 404 });
  }
  return json({ net: net });
};
export default function ControlSystemPage() {
  const { net } = useLoaderData<typeof loader>();

  return (
    <div>
      <h1>Control</h1>
      <main>
        <PetriNetProvider net={net}>
          <Outlet />
        </PetriNetProvider>
      </main>
    </div>
  );
}
;