import { Outlet, useLoaderData } from "@remix-run/react";
import { PetriNetProvider } from "~/lib/context/petrinet";
import { json, LoaderArgs } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { getNet } from "~/models/net.server";

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
  return json({ net: net, user: user });
};

export default function ControlRoute() {
  const { net, user } = useLoaderData<typeof loader>();

  return (
    <PetriNetProvider net={net} userID={user.id}>
      <Outlet />
    </PetriNetProvider>
  );
}