import { MouseEventHandler } from "react";

type ButtonBoxProps = {
  name: string
  handlers: string[]
  onClick: MouseEventHandler<HTMLButtonElement>
}

export default function ButtonBox(props: ButtonBoxProps) {
  return (
    <div className="flex items-center flex-col">
      <div className="flex items-center flex-row space-x-1">
        {props.handlers.map((route, k) => {
            return (
              <button
                onClick={props.onClick}
                key={k}
                name={route}
                className="rounded bg-slate-600 px-1 py-1 text-blue-100 hover:bg-blue-500 active:bg-blue-600"
              >
                {route}
              </button>
            );
          }
        )}
      </div>
      <h1 className="text-md font-bold">
        {props.name}
      </h1>
    </div>
  );
}

