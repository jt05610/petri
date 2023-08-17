import { NavLink } from "@remix-run/react";

export default function TransitionsIndex() {
  return (
    <div>
      <ul>
        <NavLink to={"new"}>New transition</NavLink>
      </ul>
    </div>
  );
}