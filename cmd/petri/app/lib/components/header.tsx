import { Form } from "@remix-run/react";
import { ReactNode } from "react";

type Props = {
  children?: ReactNode;
}

export default function Header({ children }: Props) {
  return (
    <header className="flex items-center justify-between bg-slate-800 p-4 text-white dark:bg-transparent">
      <div className={"flex flex-row"}>
        {children}
      </div>
      <div>
        <Form action="/logout" method="post">
          <button
            type="submit"
            className="rounded px-4 text-blue-100 dark:text-gray-200 hover:dark:text-teal-500 active:dark:text-teal-600 hover:text-blue-500 active:text-blue-600"
          >
            Logout
          </button>
        </Form>
      </div>
    </header>
  );
}