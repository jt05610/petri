import { Outlet, useLoaderData } from "@remix-run/react";
import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import { listDevices } from "~/models/device.server";
import { getUserById } from "~/models/user.server";

export const loader = async ({ request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const deviceListItems = await listDevices({ authorID });

  return json({ deviceListItems, user });
};

export default function DeviceRoute() {
  const { deviceListItems } = useLoaderData<typeof loader>();
  return (
    <div>
      <h1>Device</h1>
      <main>
        <div className={"flex w-full min-h-screen flex-row"}>
          <div className={"flex w-3/10 min-h-full flex-col bg-slate-500"}>
            <h2>Device List</h2>
            <ul>
              {deviceListItems.map((deviceListItem) => (
                <li key={deviceListItem.id}>
                  <a href={`/device/${deviceListItem.id}`}>
                    {deviceListItem.name}
                  </a>
                </li>
              ))
              }
            </ul>
          </div>
          <div className={"flex w-full min-h-screen flex-col bg-white"}>
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  );
}