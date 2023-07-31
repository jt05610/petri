import { NavLink } from "@remix-run/react";
import Dropdown from "~/lib/components/dropdown";
import { useContext } from "react";
import { NetListContext } from "~/context/netList";

export default function DesignIndexPage() {
  const items = useContext(NetListContext);
  return (
    <div className={`h-full w-full p-2`}>
      <h2 className={"text-2xl font-bold"}>No design loaded</h2>
      <div className={"flex flex-row p-3 gap-4 place-content-center items-center w-full h-full"}>
        <div className={"flex flex-col"}>
          <NavLink
            className={`grow-0 rounded bg-slate-600 px-4 py-2 text-blue-100 hover:bg-blue-500 active:bg-blue-600`}
            to={"new"}
          >
            New
          </NavLink>
        </div>

        <div className={"flex flex-col"}>
          <Dropdown current={"Edit"} items={items.map((item) => {
            return { dest: `${item.id}`, text: item.name };
          })} />
        </div>

      </div>
    </div>
  );
}