import { Outlet, useLoaderData } from "@remix-run/react";
import { UserContext } from "~/context/user";
import { SpaceContext } from "~/context/space";

import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import { getNetListItems } from "~/models/net.server";
import { getUserById } from "~/models/user.server";
import Header from "~/lib/components/header";
import { NetListContext } from "~/context/netList";
import { NetProvider } from "~/context/net";

export const loader = async ({ request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const netListItems = await getNetListItems({ authorID });
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }

  return json({ netListItems: netListItems, user });
};

export default function DesignPage() {
  const data = useLoaderData<typeof loader>();
  return (
    <SpaceContext.Provider value={"Design"}>
      <UserContext.Provider value={data.user}>
        <NetListContext.Provider value={data.netListItems}>
            <div className="flex h-full w-full min-h-screen flex-col">
              <Header />
              <main className="flex h-full w-full bg-white">
                <div className={`flex h-full w-full items-center justify-items-center`}>
                  <Outlet />
                </div>
              </main>
            </div>
        </NetListContext.Provider>
      </UserContext.Provider>
    </SpaceContext.Provider>
  );
}