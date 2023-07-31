import { useContext } from "react";
import { Form } from "@remix-run/react";
import Dropdown from "~/lib/components/dropdown";
import { SpaceContext } from "~/context/space";
export default function Header() {
  const space = useContext(SpaceContext);
  return (
    <header className="flex items-center justify-between bg-slate-800 p-4 text-white">
      <div className={"flex flex-row"}>
        <
          Dropdown
          current={space}
          items={[
            { dest: "/design", text: "Design" },
            { dest: "/control", text: "Control" },
            { dest: "/nets", text: "Nets" }
          ]}
        />
      </div>
      <div>
        <Form action="/logout" method="post">
          <button
            type="submit"
            className="rounded bg-slate-600 px-4 py-2 text-blue-100 hover:bg-blue-500 active:bg-blue-600"
          >
            Logout
          </button>
        </Form>
      </div>
    </header>

  );
}