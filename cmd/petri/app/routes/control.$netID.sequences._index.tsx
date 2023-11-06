import { Link } from "@remix-run/react";

export default function SequenceIndexPage() {
  return (
    <Link to={"new"}>
      <button className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
        New Sequence
      </button>
    </Link>
  );
}
