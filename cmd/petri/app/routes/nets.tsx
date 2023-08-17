import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { Link, NavLink, Outlet, useLoaderData } from "@remix-run/react";
import { getNetListItems } from "~/models/net.server";
import { requireUserId } from "~/session.server";
import { getUserById } from "~/models/user.server";
import Header from "~/lib/components/header";
import { Bars3Icon, ChevronDoubleRightIcon, PlusIcon } from "@heroicons/react/24/outline";
import { ReactNode, useState } from "react";

export const loader = async ({ request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const netLIstItems = await getNetListItems({ authorID });
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  return json({ netListItems: netLIstItems, user });
};

type SidebarProps = {
  children?: ReactNode;
  visible: boolean;
}

export function Sidebar({ children, visible }: SidebarProps) {

  return (
    <aside
      className={`flex flex-col w-64 bg-white dark:bg-gray-800 border-r dark:border-gray-600 ${visible ? "" : "hidden"}`}>
      {children}
    </aside>
  );
}

export default function NetsPage() {
  const [menuVisible, setMenuVisible] = useState(false);
  const data = useLoaderData<typeof loader>();

  async function toggleMenu() {
    setMenuVisible(!menuVisible);
  }

  return (
    <div className="flex h-full min-h-screen flex-col">
      <Header>
        <button
          className={"flex flex-row justify-center items-center text-center dark:text-gray-200 hover:dark:text-teal-500 active:dark:text-teal-600 hover:text-blue-500 active:text-blue-600"}
          onClick={toggleMenu}
        >
          <Bars3Icon className="w-6 h-6" />
        </button>
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
          <h2 className="text-lg font-bold p-4">Nets</h2>
          <hr />
          {data.netListItems.length === 0 ? (
            <p className="p-4">No notes yet</p>
          ) : (
            <ol>
              {data.netListItems.map((net) => (
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
