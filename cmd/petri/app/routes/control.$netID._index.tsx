import { PetriNetContext, useSocket } from "~/context";
import { useContextSelector } from "use-context-selector";
import type { Event } from "@prisma/client";
import { useEffect, useState } from "react";

// A component that takes a net instance and returns a component that has a small form for sending each event associated with a device.
type DeviceControlPanelProps = {
  name: string,
  id: string,
}

export function DeviceControlPanel(props: DeviceControlPanelProps) {
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);
  const marking = useContextSelector(PetriNetContext, (context) => context?.marking);
  const setMarking = useContextSelector(PetriNetContext, (context) => context?.setMarking);
  const socket = useSocket();
  const [enabledEvents, setEnabledEvents] = useState<{ [eventID: string]: boolean }>({} as { [eventID: string]: boolean });

  const handleUpdateEvents = (data: { [eventID: string]: boolean }) => {
    setEnabledEvents(data);
  };

  useEffect(() => {
    if (!petriNet || !marking) return;
    const enabledEvents = petriNet.allEvents(marking).reduce((acc, event) => ({ ...acc, [event.id]: petriNet.eventEnabled(marking, event.id) }), {});
    setEnabledEvents(enabledEvents);
    console.log(`Device ${props.id} enabled events`, enabledEvents)
  }, [petriNet, marking]);

  useEffect(() => {
    console.log(`Device ${props.id} marking change`, marking);
  }, [marking, props.id]);

  function handleEvent(event: Pick<Event, "id" | "name">, data: FormData) {
    if (!petriNet || !marking || !setMarking) return;
    if (!petriNet.eventEnabled(marking, event.id)) {
      console.log("event not enabled");
      return;
    }
    const newMarking = petriNet.handleEvent(marking, event.id);
    socket?.emit("event", { data, deviceID: petriNet?.net.device!.instances![0].id, command: event.name });
    setMarking(newMarking);
    handleUpdateEvents(petriNet.allEvents(newMarking).reduce((acc, event) => ({ ...acc, [event.id]: petriNet.eventEnabled(newMarking, event.id) }), {}));
  }

  return (
    <div>
      <h1>{props.name}</h1>
      <p>Marking: {JSON.stringify(marking!, null, 2)}</p>
      {petriNet!.allEvents(marking!).map((event) => {
        return (
          <div key={event.id}>
            <h2 className={"text-xl font-bold"}>{event.name}</h2>
            <p>{event.id}</p>
            <form onSubmit={(e) => {
              e.preventDefault();
              handleEvent(event, new FormData(e.target as HTMLFormElement));
            }}>
              {event.fields!.map((field, i) => {
                return (
                  <div key={i}>
                    <label htmlFor={field.name}>{field.name}</label>
                    <input type={field.type} name={field.name} />
                  </div>
                );
              })}
              <div
                className={`flex rounded-full px-2 py-1
                bg-${event.enabled ? "fuchsia-800" : "slate-900"} 
                `}
              >
                <button
                  disabled={!enabledEvents[event.id]}
                  type="submit"
                >{enabledEvents[event.id] ? "Enabled" : "Disabled"}</button>
              </div>
            </form>
          </div>
        );
      })
      }
    </div>
  );
}

export default function ControlIndex() {
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);

  return (
    <div>
      {petriNet && petriNet.net.device && petriNet.net.device.instances && petriNet.net.device.instances.map((instance) => {
        return (
          <DeviceControlPanel
            key={instance.id}
            name={instance.name}
            id={instance.id}
          />
        );
      })
      }
    </div>
  );
};