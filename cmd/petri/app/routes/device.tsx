import { Link, NavLink, Outlet, useLoaderData } from "@remix-run/react";
import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import { listDevices } from "~/models/device.server";
import { getUserById } from "~/models/user.server";
import { useState } from "react";
import Header from "~/lib/components/header";
import {
  Bars3Icon,
  ChevronDoubleRightIcon,
  FilmIcon,
  PlusIcon,
  RectangleStackIcon,
  ServerStackIcon
} from "@heroicons/react/24/outline";
import { Sidebar } from "~/routes/nets";

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
  const [menuVisible, setMenuVisible] = useState(false);
  const { deviceListItems } = useLoaderData<typeof loader>();

  async function toggleMenu() {
    setMenuVisible(!menuVisible);
  }

  return (
    <div className="flex h-full min-h-screen flex-col">
      <Header>
        <div className={"flex flex-row space-x-3 text-center dark:text-gray-200 "}>
          <button
            className={"flex flex-row justify-center items-center hover:dark:text-teal-500 active:dark:text-teal-600 hover:text-blue-500 active:text-blue-600"}
            onClick={toggleMenu}
          >
            <Bars3Icon className="w-6 h-6" />
          </button>
          <NavLink
            className={"hover:dark:text-teal-500 active:dark:text-teal-600 hover:text-blue-500 active:text-blue-600"}
            to={"/nets"}
          >
            <RectangleStackIcon className="w-6 h-6" />
          </NavLink>
          <NavLink
            className={"hover:dark:text-teal-500 active:dark:text-teal-600 hover:text-blue-500 active:text-blue-600"}
            to={"/device"}>
            <ServerStackIcon className="w-6 h-6" />
          </NavLink>
          <NavLink
            className={"hover:dark:text-teal-500 active:dark:text-teal-600 hover:text-blue-500 active:text-blue-600"}
            to={"/control"}>
            <FilmIcon className="w-6 h-6" />
          </NavLink>

        </div>
      </Header>
      <main className="flex h-full">
        <Sidebar visible={menuVisible}>
          <h2 className="text-lg font-bold p-4">Actions</h2>
          <hr />
          <Link to="new"
                className="block p-4 text-md dark:text-gray-200 hover:dark:text-teal-500 active:dark:text-teal-600 hover:text-blue-500 active:text-blue-600">
            <PlusIcon className="w-6 h-6 inline-block" /> New net
          </Link>
          <Link to="/device/new"
                className="block p-4 text-md dark:text-gray-200 hover:dark:text-teal-500 active:dark:text-teal-600 hover:text-blue-500 active:text-blue-600">
            <PlusIcon className="w-6 h-6 inline-block" /> New device
          </Link>
          <h2 className="text-lg font-bold p-4">Devices</h2>
          <hr />
          {deviceListItems.length === 0 ? (
            <p className="p-4">No notes yet</p>
          ) : (
            <ol>
              {deviceListItems.map((net) => (
                <li key={net.id}>
                  <NavLink
                    className={({ isActive }) =>
                      `block border-b p-4 text-md dark:text-gray-200 hover:dark:text-teal-500 active:dark:text-teal-600 hover:text-blue-500 active:text-blue-600 ${isActive ? "bg-white dark:bg-slate-900" : ""}`
                    }
                    to={net.id}
                  >
                    <ChevronDoubleRightIcon className="w-6 h-6 inline-block" /> {net.name}
                  </NavLink>
                </li>
              ))}
            </ol>
          )}
        </Sidebar>
        <div className="flex-1 p-6">
          <Outlet />
        </div>
      </main>
    </div>
  );
}