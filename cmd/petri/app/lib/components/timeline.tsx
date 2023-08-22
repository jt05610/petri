import { PetriNetContext } from "~/context";
import { useContextSelector } from "use-context-selector";
import type {
  RunInputDisplay
} from "~/models/net.run.server";
import { RecordRunGridView } from "~/lib/components/displayGrid";


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

export default function Timeline({sequence}: TimelineProps) {
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);

  // when sequence gets updated, rerender

  return (
    <div className={"w-full h-3/10 bottom-0 space-x-2"}>
      <h4>Timeline</h4>
      <button
        type="submit"
        className={`px-2 py-1 text-white hover:text-sky-500 dark:hover:text-sky-400`}
        onClick={async (e) => {
          e.preventDefault();
          await fetch("sequences/new", {
            method: "POST",
            headers: {
              "Content-Type": "application/json"
            },
            body: JSON.stringify(sequence)
          }).then((res) => res.json()).then((res) => {
            console.log("response", res);
          });
        }
        }
      >
        Save
      </button>
      {
        sequence && <RunView
          minCols={10}
          minRows={2}
          deviceNames={petriNet?.devices.map((d) => d.name) ?? []}
          sequence={sequence}
        />
      }
    </div>
  )
    ;
}