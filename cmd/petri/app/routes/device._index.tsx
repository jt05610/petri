import { NavLink } from "@remix-run/react";


export default function ControlIndexRoute() {
  return (
    <div>
      <ul>
        <NavLink to={"new"}>New device</NavLink>
        <NavLink to={"operate"}>Operate device</NavLink>
      </ul>
    </div>
  );
}