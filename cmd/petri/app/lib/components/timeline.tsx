import { useContextSelector } from "use-context-selector";
import type {
  RunInputDisplay
} from "~/models/net.run.server";
import { RecordRunGridView } from "~/lib/components/displayGrid";
import { useState } from "react";
import { PetriNetContext } from "~/lib/context/petrinet";


type RunViewProps = {
  minCols: number
  minRows: number
  deviceNames: string[]
  sequence: RunInputDisplay
}

export function RunView({ minCols, minRows, deviceNames, sequence }: RunViewProps) {
  return (
    <div className="mt-4 -mb-3">
      <RecordRunGridView nCols={minCols} nRows={minRows} deviceNames={deviceNames} sequence={sequence} />
    </div>
  );
}

type TimelineProps = {
  sequence: RunInputDisplay
}

export default function Timeline({ sequence }: TimelineProps) {
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);
  const [name, setName] = useState("");


  async function handleSubmit() {
    await fetch("sequences/new", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ ...sequence, name: name })
    }).then((res) => res.json()).then((res) => {
      console.log("response", res);
    });
  }

  return (
    <div className={"w-full h-3/10 bottom-0 space-x-2"}>
      <h4>Timeline</h4>
      <label
        htmlFor={"name"}
      >
        Name
        <input
          type={"text"}
          value={name}
          className={"text-md rounded-full dark:bg-slate-700 px-2 py-1"}
          onChange={(e) => setName(e.target.value)}
        />
      </label>
      <button
        type="submit"
        className={`px-2 py-1 text-white hover:text-sky-500 dark:hover:text-sky-400`}
        onClick={handleSubmit}
      >
        Save
      </button>
      {
        sequence && <RunView
          minCols={10}
          minRows={2}
          deviceNames={petriNet?.net.devices.map((d) => d.name) ?? []}
          sequence={sequence}
        />
      }
    </div>
  );
}