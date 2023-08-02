import { NavLink } from "@remix-run/react";

export default function ArcsIndex() {
  return (
    <div>
      <ul>
        <NavLink to={"new"}>New arc</NavLink>
      </ul>
    </div>
  );
}