import type { PetriNet } from "~/util/petrinet";
import { useContextSelector } from "use-context-selector";
import { PetriNetContext, SocketContext } from "~/context";
import { useEffect, useState } from "react";
import { MarkedNet } from "~/lib/components/markedNet";
import type { Event } from "@prisma/client";
import { parse } from "@conform-to/zod";
import { z } from "zod";
import type { Command } from "../../../server/command";

type SystemControlProps = {
  net: PetriNet
}

export const makeZodSchema = (fields: {
  name: string,
  type: "string" | "number" | "boolean" | string
}[]) => {
  return fields.reduce((acc, field) => {
    switch (field.type) {
      case "string":
        return acc.extend({ [field.name]: z.string() });
      case "number":
        return acc.extend({
          [field.name]: z.preprocess(
            (a) => parseInt(z.string().parse(a), 10),
            z.number()
          )
        });
      case "boolean":
        return acc.extend({
          [field.name]: z.preprocess(
            (a) => z.string().parse(a) === "true",
            z.boolean()
          )
        });
      default:
        return acc.extend({ [field.name]: z.any() });
    }
  }, z.object({}));
};

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

  function handleEvent(event: Pick<Event, "id" | "name">, deviceID: string, data: any) {
    if (!selectedInstances[deviceID]) {
      console.log("no instance selected");
      return;
    }
    console.log("handle event", event, deviceID, data);
    if (!props.net || !marking || !setMarking) return;
    if (!props.net.eventEnabled(marking, event.id)) {
      console.log("event not enabled");
      return;
    }
    const newMarking = props.net.handleEvent(marking, event.id);
    const instance = selectedInstances[deviceID];
    socket?.emit("command", { data, input: marking, output: newMarking, deviceID: instance, command: event.name.replace(/\s/g, "_").toLowerCase() } as Command);
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
              <select
                defaultValue={""}
                className={"rounded-full p-2"}
                onChange={(e) => {
                  const instance = instances.find((instance) => instance.id === e.target.value);
                  if (!instance) return;
                  handleSelectChanged(id, instance.id);
                  const data = { data: {}, deviceID: instance.id, command: "get" };
                  console.log("systemControl sent", data);
                  socket?.emit("state", data);
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
                    <div
                      className={"flex flex-col space-y-2 border-md rounded-lg p-2"}
                      key={event.id}
                    >
                      <h2 className={"text-xl font-bold"}>{event.name}</h2>
                      <form onSubmit={(e) => {
                        const formData = new FormData(e.target as HTMLFormElement);
                        const data = parse(formData, {
                          schema: makeZodSchema(event.fields)
                        });
                        e.preventDefault();
                        if (!data.value) {
                          alert(data.error);
                        }
                        handleEvent(event, id, data.value);
                      }}>
                        {event.fields!.map((field, i) => {
                          return (
                            <div
                              key={i}
                              className={"m-2 flex flex-row space-x-2 items-center"}
                            >
                              <label
                                className={"w-1/2 text-right font-medium"}
                                htmlFor={field.name}
                              >
                                {field.name}
                              </label>
                              <input
                                className={"rounded-full p-2 bg-transparent text-inherit w-1/2"}
                                type={field.type} name={field.name}
                              />
                            </div>
                          );
                        })}
                        <button
                          className={`flex rounded-full px-2 py-1 font-medium text-white flex-grow-0 flex-shrink ${event.enabled ? "bg-green-700" : "bg-slate-900"} `}
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