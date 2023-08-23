import { Menu, Transition } from "@headlessui/react";
import { Fragment } from "react";
import type { ReactNode } from "react";
import { ChevronDownIcon } from "@heroicons/react/24/outline";

type ComponentDropdownProps = {
  title: string
  children: ReactNode;
}

type DropdownItemProps = {
  key: string
  children: ReactNode;
  isActive: boolean
}

export function DropdownItem({ key, children, isActive }: DropdownItemProps) {
  return (
    <Menu.Item key={key}>
      {({ active }) => (
        <div
          className={`${
            active ? "bg-violet-500 text-white" : "text-gray-900"
          } group flex w-full items-center rounded-md px-2 py-2 text-sm ${isActive ? "bg-white" : ""}`}
        >
          {children}
        </div>
      )}
    </Menu.Item>
  );
}


export default function Dropdown({ title, children }: ComponentDropdownProps) {
  return (
    <Menu as="div" className="relative inline-block text-left">
      <Menu.Button
        className="inline-flex w-full justify-center rounded-full bg-black bg-opacity-20 px-4 py-2 text-sm font-medium text-white hover:bg-opacity-30 focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75">
        {title}
        <ChevronDownIcon
          className="ml-2 -mr-1 h-5 w-5 text-violet-200 hover:text-teal-100"
          aria-hidden="true"
        />
      </Menu.Button>
      <Transition
        as={Fragment}
        enter="transition ease-out duration-100"
        enterFrom="transform opacity-0 scale-95"
        enterTo="transform opacity-100 scale-100"
        leave="transition ease-in duration-75"
        leaveFrom="transform opacity-100 scale-100"
        leaveTo="transform opacity-0 scale-95"
      >
        <Menu.Items
          className="absolute right-0 mt-2 w-56 origin-top-right divide-y divide-gray-100 rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
          <div className="px-1 py-1 ">
            {children}
          </div>
        </Menu.Items>
      </Transition>
    </Menu>
  );
}

