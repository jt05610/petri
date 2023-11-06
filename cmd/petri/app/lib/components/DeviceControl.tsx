import type { DeviceWithEvents } from "~/util/petrinet";
import { EyeIcon, EyeSlashIcon } from "@heroicons/react/24/solid";
import { validEntry } from "~/util/parameters";
import type { FormEvent } from "react";
import { useState } from "react";
import type { ActionInputDisplay, ConstantInputDisplay } from "~/models/net.run";
import cloneDeep from "lodash/cloneDeep";
import { PetriNetActionType, PetriNetContext } from "~/lib/context/petrinet";
import { RecordRunContext, RunActionType } from "~/context";
import { useContextSelector } from "use-context-selector";
import type { EventDetails } from "~/models/net.transition.event.server";
import type { Event } from "@prisma/client";
import { ParserContext } from "~/lib/context/ParserContext";

type EventData = {
  id: Event["id"],
  name: Event["name"]
  fields: {
    id: string,
    name: string,
    type: EventDetails["fields"][0]["type"]
  }[]
}

export function DeviceControl(
  {
    id,
    name,
    events
  }: DeviceWithEvents) {
  const [hidden, setHidden] = useState<boolean>(false);
  const net = useContextSelector(PetriNetContext, (context) => context);
  const dispatch = useContextSelector(RecordRunContext, (context) => context?.dispatch);
  const parameters = useContextSelector(ParserContext, (context) => context ? {
    scope: context.state.scope,
    dispatch: context.dispatch
  } : {
    scope: undefined,
    dispatch: undefined
  });

  function toggleHidden() {
    setHidden(!hidden);
  }

  function handleEvent(event: EventData, deviceID: string, data: any) {
    console.log("handle event", event, deviceID, data);
    if (!net) return;
    if (!net.petriNet.net.eventEnabled(net.petriNet.marking, event.id)) {
      console.log("event not enabled");
      return;
    }
    // map the message fieldNames to the event's field ids
    const constants: ConstantInputDisplay[] = event.fields.map((field) => {
      if (!parameters) {
        return {
          fieldID: field.id,
          fieldName: field.name,
          constant: false,
          value: data[field.name] ?? ""
        };
      }
      if (!parameters.scope) {
        return {
          fieldID: field.id,
          fieldName: field.name,
          constant: false,
          value: data[field.name] ?? ""
        };
      }
      if (!validEntry(parameters.scope, data[field.name] ?? "")) {
        return {
          fieldID: field.id,
          fieldName: field.name,
          constant: false,
          value: ""
        };
      }
      return {
        fieldID: field.id,
        fieldName: field.name,
        constant: false,
        value: data[field.name] ?? ""
      };
    });

    const oldMarking = cloneDeep(net.petriNet.marking);
    const newMarking = net.petriNet.net.handleEvent(net.petriNet.marking, event.id);

    net.dispatch({
      type: PetriNetActionType.UpdateMarking,
      payload: newMarking
    });

    // any data with the message is assumed to be a constant for this event
    const actionInput: ActionInputDisplay = {
      eventName: event.name,
      eventID: event.id,
      deviceId: deviceID,
      eventFields: event.fields,
      input: oldMarking,
      output: newMarking,
      constants
    };

    if (!dispatch) return;
    console.log("dispatching", actionInput);
    if (!parameters.scope) return;
    dispatch({
      type: RunActionType.ActionAdded,
      payload: {
        input: actionInput,
        scope: parameters.scope
      }
    });
  }

  function handleSubmit(event: EventData) {
    return (e: FormEvent<HTMLFormElement>) => {
      const formData = new FormData(e.target as HTMLFormElement);
      e.preventDefault();
      const data = Object.fromEntries(formData.entries());
      if (!parameters) return;
      if (!parameters.scope) return;
      for (const field of event.fields!) {
        if (!validEntry(parameters.scope, data[field.name] as string)) {
          return;
        }
      }
      handleEvent(event, id, data);
    };
  }

  return (
    <div
      key={`${id}_${name}`}
      className={`flex flex-col border-2 dark:border-gray-50 dark:border-opacity-30 rounded-lg p-2  w-96 ${hidden ? "h-fit" : "h-96"}`}
    >
      <div className={"flex flex-row space-x-2 justify-between sticky py-1 px-2"}>
        <h2 className={"text-xl font-bold"}>{name}</h2>
        <button
          onClick={toggleHidden}>
          {hidden ?
            <EyeIcon className={"w-6 h-6"} /> :
            <EyeSlashIcon className={"w-6 h-6"} />}
        </button>
      </div>
      <div className={"overflow-auto"}>
        {!hidden && events.map((event, index) => {
            return (
              <div
                className={"flex flex-col space-y-2 border-md rounded-lg p-2"}
                key={`${event.id}_${index}`}
              >
                <h2 className={"text-xl font-bold"}>{event.name}</h2>
                <form onSubmit={handleSubmit(event)}>
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
                          name={field.name}
                        />
                      </div>
                    );
                  })}
                  <button
                    className={`flex rounded-full px-2 py-1 font-medium text-white flex-grow-0 flex-shrink ${event.enabled ? "bg-green-700" : "bg-slate-900"} `}
                    disabled={net!.petriNet.enabledEvents[event.id] !== null ? !net!.petriNet.enabledEvents[event.id] : true}
                    type="submit"
                  >{net!.petriNet.enabledEvents[event.id] ? "Send" : "Disabled"}</button>
                </form>
              </div>
            );
          }
        )}
      </div>
    </div>
  );
}