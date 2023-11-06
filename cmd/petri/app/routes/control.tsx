import { Link, NavLink, Outlet, useLoaderData } from "@remix-run/react";
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
import type { ReactNode } from "react";
import { useState } from "react";
import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import { getUserById } from "~/models/user.server";
import { getNetsWithDevice, getNetsWithEvents } from "~/models/net.server";
import { createContext } from "use-context-selector";

export const loader = async ({ request }: LoaderArgs) => {
  const userID = await requireUserId(request);
  const user = await getUserById(userID);
  if (!user) {
    throw new Error("User not found");
  }
  const systemNets = await getNetsWithEvents({ authorID: userID });
  const deviceNets = await getNetsWithDevice({ authorID: userID });
  return json({ systemNets, deviceNets });
};

export const ControlContext = createContext<{
  systemNets: NetListItem[],
  deviceNets: NetListItem[]
} | null>(null);

type NetListItem = {
  id: string;
  name: string;
  description: string;
}

type ControlProviderProps = {
  children?: ReactNode;
  systemNets: NetListItem[];
  deviceNets: NetListItem[];
}

export function ControlProvider({ children, systemNets, deviceNets }: ControlProviderProps) {
  return <ControlContext.Provider value={{ systemNets, deviceNets }}>{children}</ControlContext.Provider>;
}

export default function ControlRoute() {
  const [menuVisible, setMenuVisible] = useState(false);
  const { systemNets, deviceNets } = useLoaderData<typeof loader>();

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
          <h2 className="text-lg font-bold p-4">System Nets</h2>
          <hr />
          {systemNets.length === 0 ? (
            <p className="p-4">No system nets yet.</p>
          ) : (
            <ol>
              {systemNets.map((net) => (
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
          <h2 className="text-lg font-bold p-4">Device nets</h2>
          <hr />
          {deviceNets.length === 0 ? (
            <p className="p-4">No device nets yet.</p>
          ) : (
            <ol>
              {deviceNets.map((net) => (
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
          <ControlProvider systemNets={systemNets} deviceNets={deviceNets}>
            <Outlet />
          </ControlProvider>
        </div>
      </main>
    </div>
  );
}