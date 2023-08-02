import { NavLink } from "@remix-run/react";

export default function PlacesIndex() {
  return (
    <div>
      <ul>
        <NavLink to={"new"}>New place</NavLink>
      </ul>
    </div>
  );
}