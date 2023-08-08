import { Outlet, useLoaderData } from "@remix-run/react";
import { PetriNetProvider, SocketContext, SocketProvider } from "~/context";
import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { getNet } from "~/models/net.server";
import { useContextSelector } from "use-context-selector";

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

export default function PlaySequence() {
  const { net } = useLoaderData<typeof loader>();
  const socket = useContextSelector(SocketContext, (context) => context);
  return (
    <div className={"flex flex-col h-screen w-full items-center justify-items-center"}>
      <SocketProvider socket={socket}>
        <PetriNetProvider net={net}>
          <Outlet />
        </PetriNetProvider>
      </SocketProvider>
    </div>
  );
};