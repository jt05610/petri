import { useContextSelector } from "use-context-selector";
import { RunContext } from "~/lib/context/run";
import RunGridView from "~/lib/components/displayGrid";
import type { RunDetails } from "~/models/net.run.server";
import type { Device } from "@prisma/client";
import { MutableRefObject } from "react";
import { ParameterWithValue } from "~/lib/context/session";

type RunViewProps = {
  minCols: number
  minRows: number
  sequence: RunDetails
  paramRef?: MutableRefObject<Record<string, Record<number, Record<string, ParameterWithValue>>>>
  devices: Pick<Device, "id" | "name">[]
}

export function RunView({ minCols, minRows, devices, sequence, paramRef }: RunViewProps) {
  return (
    <div className="mt-4 -mb-3">
      <RunGridView nCols={minCols} nRows={minRows} deviceNames={devices.map((d) => d.name)} sequence={sequence} paramRef={paramRef} />
    </div>
  );
}

type PlayerProps = {
  devices: Pick<Device, "id" | "name">[]
  paramRef: MutableRefObject<Record<string, Record<number, Record<string, ParameterWithValue>>>>
}

export default function Player({ devices, paramRef }: PlayerProps) {
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
      {
        sequence && <RunView
          minCols={20}
          minRows={10}
          devices={devices}
          sequence={sequence}
          paramRef={paramRef}
        />
      }
    </div>
  )
    ;
}