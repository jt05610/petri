import { NavLink } from "@remix-run/react";

export default function InstanceIndex() {
  return (
    <div>
      <ul>
        <NavLink to={"new"}>New Instance</NavLink>
      </ul>
    </div>
  );
}