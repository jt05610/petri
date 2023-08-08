import type { ReactNode } from "react";
import { PetriNetContext, RecordRunContext, RunActionType, SocketContext } from "~/context";
import { useContextSelector } from "use-context-selector";
import type {
  ConstantInputDisplay,
  ActionInputDisplay,
  RunInputDisplay
} from "~/models/net.run.server";
import { useEffect, useRef } from "react";
import type { Marking } from "~/util/petrinet";

export default function Player() {
  const socket = useContextSelector(SocketContext, (context) => context);
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);
  const petriNetRef = useRef<typeof petriNet>();
  const dispatch = useContextSelector(RecordRunContext, (context) => context?.dispatch);
  const dispatchRef = useRef<typeof dispatch>();
  const sequence = useContextSelector(RecordRunContext, (context) => context?.run);
  const seqRef = useRef<typeof sequence>();

  useEffect(() => {
    console.log("rendering timeline");
  }, []);
  useEffect(() => {
    if (!dispatch) return;
    dispatchRef.current = dispatch;
  }, [dispatch]);

  useEffect(() => {
    if (!sequence) return;
    seqRef.current = sequence;
  }, [sequence]);

  useEffect(() => {
    if (!petriNet) return;
    petriNetRef.current = petriNet;
  }, [petriNet]);

  useEffect(() => {
    if (!socket) return;
    socket.on("command", (data: {
      deviceID: string,
      event: string,
      input: Marking,
      output: Marking,
      message: {
        [fieldNme: string]: string
      }
    }) => {
      console.log("command", data);
      // replace whitespace in event name with underscores and make lowercase
      const transformedEventName = (event: string) => event.replace(/\s/g, "_").toLowerCase();
      const event = petriNetRef.current!.events.find((e) => data.event === transformedEventName(e.name));
      // make human-readable timestamp in format DD MMM YYYY HH:MM:SS
      if (!event) {
        console.log("event not found");
        return;
      }
      // map the message fieldNames to the event's field ids
      const constants: ConstantInputDisplay[] = event.fields.map((field) => {
        return {
          fieldID: field.id,
          fieldName: field.name,
          constant: false,
          value: data.message[field.name]
        };
      });

      // any data with the message is assumed to be a constant for this event
      const actionInput: ActionInputDisplay = {
        eventName: event.name,
        eventID: event.id,
        deviceId: petriNetRef.current!.instanceOf(data.deviceID),
        input: data.input,
        output: data.output,
        constants
      };
      console.log("dispatching", actionInput);

      dispatchRef.current!({
        type: RunActionType.ActionAdded,
        payload: actionInput
      });
    });
  }, [socket]);

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