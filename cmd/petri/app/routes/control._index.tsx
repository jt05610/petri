import { NavLink } from "@remix-run/react";

export default function ControlIndexRoute() {
  return (
    <div>
      <ul>
        <NavLink to={"new"}>New protocol</NavLink>
        <NavLink to={"run"}>Run protocol</NavLink>
        <NavLink to={"edit"}>Edit a protocol</NavLink>
      </ul>
      <h3>Operate</h3>
      <ul>
        <NavLink to={"devices"}>Devices</NavLink>
      </ul>
    </div>
  );
}