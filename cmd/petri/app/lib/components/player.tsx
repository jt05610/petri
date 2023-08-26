import {
  PetriNetContext
} from "~/lib/context/petrinet";
import { useContextSelector } from "use-context-selector";
import { RunContext } from "~/lib/context/run";
import RunGridView from "~/lib/components/displayGrid";
import type { RunDetails } from "~/models/net.run.server";

type RunViewProps = {
  minCols: number
  minRows: number
  deviceNames: string[]
  sequence: RunDetails
}

export function RunView({ minCols, minRows, deviceNames, sequence }: RunViewProps) {
  return (
    <div className="mt-4 -mb-3">
      <RunGridView nCols={minCols} nRows={minRows} deviceNames={deviceNames} sequence={sequence} />
    </div>
  );
}

export default function Player() {
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet.petriNet);
  const sequence = useContextSelector(RunContext, (context) => context?.run);
  /*
    useEffect(() => {
      if (!socket || !petriNet || !session || !sequence) return;
      socket.on("command", (data: {
        deviceID: string,
        command: string,
        input: Marking,
        output: Marking,
        message: {
          [fieldNme: string]: string
        }
      }) => {
        console.log("command", data);
        // replace whitespace in event name with underscores and make lowercase
        const transformedEventName = (event: string) => event.replace(/\s/g, "_").toLowerCase();
        const event = petriNet!.events.find((e) => data.command === transformedEventName(e.name));
        // make human-readable timestamp in format DD MMM YYYY HH:MM:SS
        if (!event) {
          console.log("event not found");
          return;
        }
        dispatch!({
          type: SessionActionType.ActionStarted,
          payload: {}
        });
      });
      socket.on("event", (data: {
        deviceID: string,
        event: string,
        input: Marking,
        output: Marking,
        message: {
          [fieldNme: string]: string
        }
      }) => {
        console.log("event", data);
        // replace whitespace in event name with underscores and make lowercase
        const transformedEventName = (event: string) => event.replace(/\s/g, "_").toLowerCase();
        const event = petriNet!.events.find((e) => data.event === transformedEventName(e.name));
        // make human-readable timestamp in format DD MMM YYYY HH:MM:SS
        if (!event) {
          console.log("event not found");
          return;
        }
        dispatch!({
          type: SessionActionType.ActionCompleted,
          payload: {}
        });
      });
    }, [dispatch, petriNet, sequence, session, socket]);
  */
  return (
    <div className={"w-full h-3/10 bottom-0 space-x-2"}>
      <h4>Player</h4>
      <button
        type="submit"
        className={`rounded-full px-2 py-1 text-white bg-slate-900`}
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