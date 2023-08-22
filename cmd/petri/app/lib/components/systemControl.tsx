import { useContextSelector } from "use-context-selector";
import { PetriNetContext, RecordRunContext, RunActionType } from "~/context";
import { Suspense, useEffect, useState } from "react";
import { MarkedNet } from "~/lib/components/markedNet";
import { parse } from "@conform-to/zod";
import { z } from "zod";
import type { ActionInputDisplay, ConstantInputDisplay } from "~/models/net.run.server";
import type { EventDetails } from "~/models/net.transition.event.server";
import type { Event } from "@prisma/client";
import Timeline from "~/lib/components/timeline";

export const makeZodSchema = (fields: {
  id: string
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


type EventData = {
  id: Event["id"],
  name: Event["name"]
  fields: {
    id: string,
    name: string,
    type: EventDetails["fields"][0]["type"]
  }[]
}

export function SystemControl() {
  const net = useContextSelector(PetriNetContext, (context) => context?.petriNet);
  const marking = useContextSelector(PetriNetContext, (context) => context?.marking);
  const run = useContextSelector(RecordRunContext, (context) => context?.run);
  const dispatch = useContextSelector(RecordRunContext, (context) => context?.dispatch);
  const setMarking = useContextSelector(PetriNetContext, (context) => context?.setMarking);
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
    if (!net || !marking) return;
    const enabledEvents = net.allEvents(marking).reduce((acc, event) => ({
      ...acc,
      [event.id]: net.eventEnabled(marking, event.id)
    }), {});
    setEnabledEvents(enabledEvents);
    console.log("enabled events", enabledEvents);
  }, [marking, net]);

  function handleEvent(event: EventData, deviceID: string, data: any) {
    if (!selectedInstances[deviceID]) {
      console.log("no instance selected");
      return;
    }
    console.log("handle event", event, deviceID, data);
    if (!net || !marking || !setMarking) return;
    if (!net.eventEnabled(marking, event.id)) {
      console.log("event not enabled");
      return;
    }
    // map the message fieldNames to the event's field ids
    const constants: ConstantInputDisplay[] = event.fields.map((field) => {
      return {
        fieldID: field.id,
        fieldName: field.name,
        constant: false,
        value: `${data[field.name]}` ?? ""
      };
    });

    const newMarking = net.handleEvent(marking, event.id);

    console.log(net.deviceIDFromInstanceID)
    // any data with the message is assumed to be a constant for this event
    const actionInput: ActionInputDisplay = {
      eventName: event.name,
      eventID: event.id,
      deviceId: deviceID,
      eventFields: event.fields,
      input: marking,
      output: newMarking,
      constants
    };

    if (!dispatch) return;
    console.log("dispatching", actionInput);

    dispatch({
      type: RunActionType.ActionAdded,
      payload: actionInput
    });

    setMarking(newMarking);
    handleUpdateEvents(net.allEvents(newMarking).reduce((acc, event) => ({
      ...acc,
      [event.id]: net.eventEnabled(newMarking, event.id)
    }), {}));

  }

  return (
    <div className={"flex flex-col h-screen w-full items-center justify-items-center"}>
      <div className={"h-7/10 w-full"}>
        <div className={"flex h-full w-full flex-row space-x-2 p-2"}>
          <div className={"flex w-full flex-col space-y-2"}>
            {net && net.childDeviceEvents(marking!).map(({ id, name, instances, events }) => {
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
            {marking && net && <MarkedNet marking={marking} net={net} />}
          </div>
        </div>
      </div>
      <div className={"h-3/10 w-full"}>
        <Suspense fallback={<div>Loading...</div>}>
        {marking && net && run &&
          <Timeline sequence={run} />
        }
        </Suspense>
      </div>
    </div>
  );
}