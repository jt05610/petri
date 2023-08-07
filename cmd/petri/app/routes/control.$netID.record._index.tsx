import { PetriNetContext, RecordSequenceProvider, SocketContext } from "~/context";
import { useContextSelector } from "use-context-selector";
import type { Event } from "@prisma/client";
import { useEffect, useState } from "react";
import Timeline from "~/lib/components/timeline";
import type { PetriNet } from "~/util/petrinet";
import { MarkedNet } from "~/lib/components/markedNet";
import type { LoaderArgs } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import { getUserById } from "~/models/user.server";
import invariant from "tiny-invariant";
import { addSequence, SequenceInputSchema } from "~/models/sequence.server";
import { redirect } from "@remix-run/node";


type SystemControlProps = {
  net: PetriNet
}

export function SystemControl(props: SystemControlProps) {
  const marking = useContextSelector(PetriNetContext, (context) => context?.marking);
  const setMarking = useContextSelector(PetriNetContext, (context) => context?.setMarking);
  const socket = useContextSelector(SocketContext, (context) => context);
  const [enabledEvents, setEnabledEvents] = useState<{
    [eventID: string]: boolean
  }>({} as {
    [eventID: string]: boolean
  });
  const [selectedInstances, setSelectedInstances] = useState<{
    [deviceID: string]: string
  }>({} as {
    [deviceID: string]: string
  });

  const handleUpdateEvents = (data: {
    [eventID: string]: boolean
  }) => {
    setEnabledEvents(data);
  };

  const handleSelectChanged = (deviceID: string, instanceID: string) => {
    setSelectedInstances({ ...selectedInstances, [deviceID]: instanceID });
  };

  useEffect(() => {
    if (!props.net || !marking) return;
    const enabledEvents = props.net.allEvents(marking).reduce((acc, event) => ({
      ...acc,
      [event.id]: props.net.eventEnabled(marking, event.id)
    }), {});
    setEnabledEvents(enabledEvents);
    console.log("enabled events", enabledEvents);
  }, [marking, props.net]);

  function handleEvent(event: Pick<Event, "id" | "name">, deviceID: string, data: FormData) {
    if (!selectedInstances[deviceID]) {
      console.log("no instance selected");
      return;
    }
    if (!props.net || !marking || !setMarking) return;
    if (!props.net.eventEnabled(marking, event.id)) {
      console.log("event not enabled");
      return;
    }
    const newMarking = props.net.handleEvent(marking, event.id);
    const instance = selectedInstances[deviceID];
    socket?.emit("event", { data, deviceID: instance, command: event.name });
    setMarking(newMarking);
    handleUpdateEvents(props.net.allEvents(newMarking).reduce((acc, event) => ({
      ...acc,
      [event.id]: props.net.eventEnabled(newMarking, event.id)
    }), {}));
  }

  return (
    <div className={"flex h-full w-full flex-row space-x-2 p-2"}>
      <div className={"flex w-full flex-col space-y-2"}>
        {props.net.childDeviceEvents(marking!).map(({ id, name, instances, events }) => {
          return (
            <div
              key={id}
              className={"flex flex-col space-y-2 border-md rounded-lg p-2"}
            >
              <h2 className={"text-xl font-bold"}>{name}</h2>
              <p>{id}</p>
              <select
                defaultValue={""}
                className={"rounded-full p-2"}
                onChange={(e) => {
                  const instance = instances.find((instance) => instance.id === e.target.value);
                  if (!instance) return;
                  handleSelectChanged(id, instance.id);
                  socket?.emit("event", { data: {}, deviceID: instance.id, command: "get" });
                }
                }>
                <option value={""}>Select a device</option>
                {instances?.map((instance) => {
                    return (
                      <option key={instance.id} value={instance.id}>{instance.name}</option>
                    );
                  }
                )}
              </select>
              {events.map((event) => {
                  return (
                    <div key={event.id}>
                      <h2 className={"text-xl font-bold"}>{event.name}</h2>
                      <form onSubmit={(e) => {
                        e.preventDefault();
                        handleEvent(event, id, new FormData(e.target as HTMLFormElement));
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
                }
              )}
            </div>
          );
        })
        }
      </div>
      <div
        className={"flex h-full w-full border-2 border-gray-900 rounded-xl items-center p-2 space-x-2"}
      >
        {marking && <MarkedNet marking={marking} net={props.net} />}
      </div>
    </div>

  );
}


export default function ControlIndex() {
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);
  console.log(petriNet?.toGraphViz());
  const marking = useContextSelector(PetriNetContext, (context) => context?.marking);
  return (
    <div className={"flex flex-col h-screen w-full items-center justify-items-center"}>
      <div className={"h-7/10 w-full"}>
        <SystemControl net={petriNet!} />
      </div>
      <div className={"h-3/10 w-full"}>
        {marking && petriNet &&
          <RecordSequenceProvider>
            <Timeline />
          </RecordSequenceProvider>
        }
      </div>
    </div>
  );
};