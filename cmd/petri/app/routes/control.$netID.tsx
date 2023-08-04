import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getNetWithDeviceInstances } from "~/models/net.server";
import { Outlet, useLoaderData } from "@remix-run/react";
import {
  PetriNetProvider, SocketProvider,
} from "~/context";
import { getUserById } from "~/models/user.server";
import { useEffect, useState } from "react";
import { io, Socket } from "socket.io-client";

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.netID, "netID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const net = await getNetWithDeviceInstances({ id: params.netID, authorID: authorID });
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

  useEffect(() => {
    if (!socket) return;
    socket.on("confirmation", (data) => {
      console.log(data);
    });
  }, [socket]);

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