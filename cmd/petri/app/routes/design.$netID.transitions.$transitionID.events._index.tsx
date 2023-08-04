import { NavLink } from "@remix-run/react";

export default function EventsIndex() {
  return (
    <div>
      <ul>
        <NavLink to={"new"}>New event</NavLink>
      </ul>
    </div>
  );
}