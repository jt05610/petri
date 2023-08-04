import { Outlet } from "@remix-run/react";

import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import { getNetListItems } from "~/models/net.server";
import { getUserById } from "~/models/user.server";
import Header from "~/lib/components/header";
import { NetListContext } from "~/context/netList";
import { NetProvider } from "~/context/net";



export default function DesignPage() {
  return (

    <div className="flex h-full w-full min-h-screen flex-col">
      <Header />
      <main className="flex h-full w-full bg-white">
        <div className={`flex h-full w-full items-center justify-items-center`}>
          <Outlet />
        </div>
      </main>
    </div>

  );
}