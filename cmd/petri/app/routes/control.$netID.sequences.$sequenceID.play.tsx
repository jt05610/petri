import { Outlet } from "@remix-run/react";

export default function PlaySequence() {
  return (
    <div className={"flex flex-col h-screen w-full items-center justify-items-center"}>
      <Outlet />
    </div>
  );
};