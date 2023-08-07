import { useContextSelector } from "use-context-selector";
import { PetriNetContext, SocketContext } from "~/context";
import { Suspense, useEffect, useState } from "react";
import { MarkedNet } from "~/lib/components/markedNet";
import type { PetriNet } from "~/util/petrinet";
import type { Event } from "@prisma/client";

type DeviceControlPanelProps = {
  name: string,
  id: string,
}

export function DeviceControlPanel(props: DeviceControlPanelProps) {
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);
  const socket = useContextSelector(SocketContext, (context) => context);
  const marking = useContextSelector(PetriNetContext, (context) => context?.marking);
  const setMarking = useContextSelector(PetriNetContext, (context) => context?.setMarking);
  const [enabledEvents, setEnabledEvents] = useState<{
    [eventID: string]: boolean
  }>({} as {
    [eventID: string]: boolean
  });

  const handleUpdateEvents = (data: {
    [eventID: string]: boolean
  }) => {
    setEnabledEvents(data);
  };

  useEffect(() => {
    if (!petriNet || !marking) return;
    const enabledEvents = petriNet.allEvents(marking).reduce((acc, event) => ({
      ...acc,
      [event.id]: petriNet.eventEnabled(marking, event.id)
    }), {});
    setEnabledEvents(enabledEvents);
    console.log(`Device ${props.id} enabled events`, enabledEvents);
  }, [props.id, petriNet, marking, setEnabledEvents]);

  function handleEvent(event: Pick<Event, "id" | "name">, data: FormData) {
    if (!petriNet || !marking || !setMarking) return;
    if (!petriNet.eventEnabled(marking, event.id)) {
      console.log("event not enabled");
      return;
    }
    const newMarking = petriNet.handleEvent(marking, event.id);
    socket?.emit("event", { data, deviceID: petriNet?.net.device!.instances![0].id, command: event.name });
    setMarking(newMarking);
    handleUpdateEvents(petriNet.allEvents(newMarking).reduce((acc, event) => ({
      ...acc,
      [event.id]: petriNet.eventEnabled(newMarking, event.id)
    }), {}));
  }

  return (
    <div
      className={"flex h-full flex-col space-y-2 p-2"}
    >
      <Suspense fallback={<p>Marking: {JSON.stringify(marking!, null, 2)}</p>}>
        <MarkedNet marking={marking!} net={petriNet!} />
      </Suspense>

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
              <button
                className={`flex rounded-full px-2 py-1 text-white flex-grow-0 flex-shrink ${event.enabled ? "bg-green-700" : "bg-slate-900"} `}
                disabled={enabledEvents[event.id] !== null ? !enabledEvents[event.id] : true}
                type="submit"
              >{enabledEvents[event.id] ? "Send" : "Disabled"}</button>
            </form>
          </div>
        );
      })
      }
    </div>
  );
}

type DeviceControlProps = {
  instances: {
    id: string,
    name: string,
    addr: string
  }[]
}

export function DeviceControl(props: DeviceControlProps) {
  const [selectedInstance, setInstance] = useState<{
    id: string,
    name: string,
    addr: string
  }>({
    id: "",
    name: "",
    addr: ""
  });
  const handleInstanceChanged = (data: {
    id: string,
    name: string,
    addr: string
  }) => {
    setInstance(data);
  };

  return (
    <div>
      <select
        className={"rounded-full p-2"}
        onChange={(e) => {
          const instance = props.instances.find((instance) => instance.id === e.target.value);
          if (!instance) return;
          handleInstanceChanged({ id: instance.id, name: instance.name, addr: instance.addr });
        }
        }
      >
        <option value={""}>Select a device</option>
        {props.instances?.map((instance) => {
          return (
            <option key={instance.id} value={instance.id}>{instance.name}</option>
          );
        })}
      </select>
      <DeviceControlPanel
        name={selectedInstance.name}
        id={selectedInstance.id}
      />
    </div>
  );
}
