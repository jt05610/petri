import { useContextSelector } from "use-context-selector";
import type {
  RunInputDisplay
} from "~/models/net.run";
import { RecordRunGridView } from "~/lib/components/displayGrid";
import { PetriNetContext } from "~/lib/context/petrinet";
import React from "react";


type RunViewProps = {
  minCols: number
  minRows: number
  deviceNames: string[]
  sequence: RunInputDisplay
}

export function RunView({ minCols, minRows, deviceNames, sequence }: RunViewProps) {
  return (
    <div className="mt-4 mb-3 w-full flex">
      <RecordRunGridView nCols={minCols} nRows={minRows} deviceNames={deviceNames} sequence={sequence} />
    </div>
  );
}

type TimelineProps = {
  sequence: RunInputDisplay
}

export default function Timeline({ sequence }: TimelineProps) {
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);

  async function handleSubmit() {
    await fetch("", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ actions: sequence.actions })
    });
  }

  return (
    <div className={"flex flex-col w-full"}>
      <h4>Timeline</h4>
      <button
        type="submit"
        className={`px-2 py-1 text-white hover:text-sky-500 dark:hover:text-sky-400`}
        onClick={handleSubmit}
      >
        Save
      </button>
      {
        sequence && (<RunView
          minCols={sequence.actions.length + 1}
          minRows={petriNet?.net.devices.length ?? 10}
          deviceNames={petriNet?.net.devices.map((d) => d.name) ?? []}
          sequence={sequence}
        />)
      }
    </div>
  );
}