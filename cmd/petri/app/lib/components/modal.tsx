import { XMarkIcon } from "@heroicons/react/24/outline";
import type { ReactNode } from "react";

type Props = {
  children: ReactNode;
  name: string
  isOpen: boolean;
  setIsOpen: (isOpen: boolean) => void;
}
export default function Modal({ children, isOpen, setIsOpen, name }: Props) {

  if (!isOpen) return null; // if not open, do not render modal and content

  return (
    <div className="fixed z-10 inset-0 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0 flex-grow">
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" aria-hidden="true"></div>
        <span className="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
        <div
          className="inline-block align-bottom dark:bg-gray-900/50 rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle">
          <div className="bg-teal-900/50 px-4 py-3 sm:px-6 sm:flex sm:justify-between">
            <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-teal-100" id="modal-title">
              {name}
            </h3>
            <button type="button"
                    className="mt-3 w-full inline-flex justify-center rounded-md border border-transparent shadow-sm bg-rose-500 text-base font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
                    onClick={() => setIsOpen(false)}>
              <XMarkIcon className="h-5 w-5 text-white" />
            </button>
          </div>
          <div className="flex sm:flex sm:items-start w-full">
            <div className="flex m-3 text-center w-full sm:m-2 sm:text-left focus overflow-auto flex-grow">
              {children}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}