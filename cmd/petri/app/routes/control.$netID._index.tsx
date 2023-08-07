import { NavLink } from "@remix-run/react";

export default function ControlIndexRoute() {
  return (
    <div>
      <ul>
        <NavLink to={"record"}>Record a new sequence of events</NavLink>
        <NavLink to={"edit"}>Edit an existing sequence</NavLink>
        <NavLink to={"execute"}>Execute a sequence of events</NavLink>
      </ul>
    </div>
  );
}