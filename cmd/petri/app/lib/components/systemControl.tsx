import { useContextSelector } from "use-context-selector";
import { RecordRunContext, RunActionType } from "~/context";
import { Suspense, useState } from "react";
import { MarkedNet } from "~/lib/components/markedNet";
import { parse } from "@conform-to/zod";
import { z } from "zod";
import type { ActionInputDisplay, ConstantInputDisplay } from "~/models/net.run.server";
import type { EventDetails } from "~/models/net.transition.event.server";
import type { Event } from "@prisma/client";
import Timeline from "~/lib/components/timeline";
import { PetriNetActionType, PetriNetContext } from "~/lib/context/petrinet";
import cloneDeep from "lodash/cloneDeep";
import { EyeIcon, EyeSlashIcon } from "@heroicons/react/24/solid";
import * as math from "mathjs";
import { validEntry } from "~/util/parameters";

export const makeZodSchema = (fields: {
  id: string
  name: string,
  type: "string" | "number" | "boolean" | string
}[], parser?: math.Parser) => {
  return fields.reduce((acc, field) => {
    switch (field.type) {
      case "string":
        return acc.extend({
          [field.name]: z.preprocess(
            (a) => {
              if (typeof a !== "string") return "";
              if (!parser) return z.string().parse(a);
              if (!validEntry(parser, a)) return "";
            },
            z.string()
          )
        });
      case "number":
        return acc.extend({
          [field.name]: z.preprocess(
            (a) => {
              if (typeof a !== "string") return NaN;
              if (!parser) return z.number().parse(a);
              if (!validEntry(parser, a)) return NaN;
            },
            z.number()
          )
        });
      case "boolean":
        return acc.extend({
          [field.name]: z.preprocess(
            (a) => {
              if (typeof a !== "string") return false;
              if (!parser) return z.boolean().parse(a);
              if (!validEntry(parser, a)) return false;
            },
            z.boolean(),
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

type ParameterProps = {
  setParser: (parser: math.Parser) => void
}

function ParameterEditor({setParser}: ParameterProps) {
  const [expression, setExpression] = useState<string>("")
  const [expressionResult, setExpressionResult] = useState<string>("");
  function validateExpression() {
    const valParser = math.parser()
    try {
      const result = valParser.evaluate(expression);
      setParser(valParser);
      setExpressionResult(`${result}`)
    } catch (e) {
      const error = e as Error;
      setExpressionResult(`Error: ${error.message}`);
      return;
    }
  }
  const onExpressionsChanged = (expression: string) => {
    setExpression(expression);
  }

  return (
    <div className={"flex flex-col space-2"}>
      <label
        htmlFor={"expression"}
      >
        Expression
      </label>
      <textarea
        name={"expression"}
        value={expression}
        className={"text-md rounded-md dark:bg-slate-700 px-2 py-1 w-1/2 font-mono h-fit max-h-52"}
        onChange={(e) => onExpressionsChanged(e.target.value)}
      />
      <div
        className={"text-md px-2 py-1 w-1/2 font-mono"}
      >
        {expressionResult}
      </div>
      <button
        className={"bg-blue-500 hover:bg-blue-700 text-white font-bold py-1 px-4 mt-2 rounded-full flex-grow-0 flex-shrink"}
        onClick={validateExpression}
      >
        Validate
      </button>
    </div>
  )
}

export function SystemControl(props: {}) {
  const [hiddenDevices, setHiddenDevices] = useState<string[]>([]);
  const net = useContextSelector(PetriNetContext, (context) => context);
  const run = useContextSelector(RecordRunContext, (context) => context?.run);
  const dispatch = useContextSelector(RecordRunContext, (context) => context?.dispatch);
  const [parser, setParser] = useState<math.Parser | undefined>(undefined);

  function hideDevice(deviceID: string) {
    setHiddenDevices([...hiddenDevices, deviceID]);
  }

  function showDevice(deviceID: string) {
    setHiddenDevices(hiddenDevices.filter((id) => id !== deviceID));
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
      if (!parser)  {
        return {
          fieldID: field.id,
          fieldName: field.name,
          constant: false,
          value: data[field.name] ?? ""
        };
      }
      if (!validEntry(parser, data[field.name] ?? "")) {
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

    dispatch({
      type: RunActionType.ActionAdded,
      payload: actionInput
    });
  }

  return (
    <div className={"flex flex-col w-full items-center justify-items-center"}>
      <div className={"w-full flex-col overflow-auto"}>
        <div className={"flex h-1/4 w-full flex-col space-x-2 p-2 overflow-y-scroll"}>
          <div
            className={"flex h-full w-full border-2 border-gray-900 rounded-xl items-center p-2 space-x-2"}
          >
            {net && <MarkedNet marking={net.petriNet.marking} net={net.petriNet.net} />}
          </div>
        </div>
        <ParameterEditor setParser={setParser}/>

        <div className={"flex h-2/4 flex-row flex-wrap gap-2"}>
          {net && net.petriNet.net.deviceEvents(net.petriNet.marking).map((
            {
              id,
              name,
              instances,
              events
            }, index) => {
            return (
              <div
                key={`${id}_${index}`}
                className={"flex flex-col border-2 dark:border-gray-50 dark:border-opacity-30 rounded-lg p-2 h-96 w-96 "}
              >
                <div className={"flex flex-row space-x-2 justify-between sticky py-1 px-2"}>
                  <h2 className={"text-xl font-bold"}>{name}</h2>
                  <button
                    onClick={() => {
                      if (hiddenDevices.includes(id)) {
                        showDevice(id);
                      } else {
                        hideDevice(id);
                      }
                    }}>
                    {hiddenDevices.includes(id) ?
                      <EyeIcon className={"w-6 h-6"} /> :
                      <EyeSlashIcon className={"w-6 h-6"} />}
                  </button>
                </div>
                <div className={"overflow-auto"}>

                  {!hiddenDevices.includes(id) && events.map((event, index) => {
                      return (
                        <div
                          className={"flex flex-col space-y-2 border-md rounded-lg p-2"}
                          key={`${event.id}_${index}`}
                        >
                          <h2 className={"text-xl font-bold"}>{event.name}</h2>
                          <form onSubmit={(e) => {
                            const formData = new FormData(e.target as HTMLFormElement);
                            e.preventDefault();
                            const data = Object.fromEntries(formData.entries());
                            // make sure all data validates
                            for (const field of event.fields!) {
                              if (!parser) continue;
                              if (!validEntry(parser, data[field.name] as string)) {
                                return;
                              }
                            }
                            handleEvent(event, id, data);
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
                              disabled={net.petriNet.enabledEvents[event.id] !== null ? !net.petriNet.enabledEvents[event.id] : true}
                              type="submit"
                            >{net.petriNet.enabledEvents[event.id] ? "Send" : "Disabled"}</button>
                          </form>
                        </div>
                      );
                    }
                  )}
                </div>
              </div>
            );
          })
          }
        </div>

      </div>
      <div className={"w-screen h-1/4"}>
        <Suspense fallback={<div>Loading...</div>}>
          {net && run &&
            <Timeline sequence={run} />
          }
        </Suspense>
      </div>
    </div>
  );
}