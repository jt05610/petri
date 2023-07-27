import { Link } from "@remix-run/react";

export default function NoteIndexPage() {
  return (
    <p>
      No net selected. Select a net on the left, or{" "}
      <Link to="new" className="text-blue-500 underline">
        create a new net.
      </Link>
    </p>
  );
}
