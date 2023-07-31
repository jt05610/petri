import { Menu } from "@headlessui/react";
import { Link } from "@remix-run/react";

export type DropdownItemProps = {
  dest: string;
  text: string;
}

type DropdownProps = {
  current: string;
  items: DropdownItemProps[];
}

export default function Dropdown(props: DropdownProps) {
  return (
    <Menu>
      <Menu.Button
        className={"rounded bg-slate-600 px-4 py-2 text-blue-100 hover:bg-blue-500 active:bg-blue-600"}
      >
        {props.current}
      </Menu.Button>
      <Menu.Items className={"flex space-x-2 items-center px-2 py-1"}>
        {props.items.map((item, i) => (
          <Menu.Item key={i}>
            {({ active }) => (
              <Link to={item.dest}>
                {item.text}
              </Link>
            )}
          </Menu.Item>
        ))}
      </Menu.Items>
    </Menu>
  );
}