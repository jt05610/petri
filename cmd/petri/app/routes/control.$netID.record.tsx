import type { LoaderArgs, ActionArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getNet } from "~/models/net.server";
import { Outlet, useLoaderData } from "@remix-run/react";
import {
  PetriNetProvider, SocketProvider
} from "~/context";
import { getUserById } from "~/models/user.server";
import { useEffect, useState } from "react";
import type { Socket } from "socket.io-client";
import { io } from "socket.io-client";

export const action = async ({ params, request }: ActionArgs) => {
  const userID = await requireUserId(request);
  invariant(params.netID, "netID not found");
  const user = await getUserById(userID);
  invariant(user, "User not found");
  return redirect("/control/" + params.netID);
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

  const [socket, setSocket] = useState<Socket>();
  useEffect(() => {
    const socket = io();
    setSocket(socket);
    return () => {
      socket.close();
    };
  }, []);

  return (
    <div>
      <h1>Control</h1>
      <main>
        <SocketProvider socket={socket}>
          <PetriNetProvider net={net}>
            <Outlet />
          </PetriNetProvider>
        </SocketProvider>
      </main>
    </div>
  );
}
;