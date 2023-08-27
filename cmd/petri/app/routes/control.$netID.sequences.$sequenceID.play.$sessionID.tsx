import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import type { RunSessionDetails } from "~/models/net.run.session.server";
import { getRunSession } from "~/models/net.run.session.server";
import invariant from "tiny-invariant";
import { requireUser } from "~/session.server";
import { Form, useLoaderData } from "@remix-run/react";
import Player from "~/lib/components/player";
import type { DeviceInstanceInput } from "~/models/__generated__/graphql";
import { DeviceMarkingsDocument } from "~/models/__generated__/graphql";
import type { FormEvent, ReactNode } from "react";
import { Suspense, useRef } from "react";
import { parse } from "@conform-to/zod";
import { z } from "zod";
import type { Parameter } from "~/models/net.run.session.data.server";
import { RunSessionActionType, RunSessionContext, RunSessionProvider } from "~/lib/context/session";
import { useQuery } from "@apollo/client";
import { useContextSelector } from "use-context-selector";
import type { Marking, PetriNet } from "~/util/petrinet";
import { PetriNetContext } from "~/lib/context/petrinet";
import { PlayIcon } from "@heroicons/react/24/solid";

export const loader = async ({ params, request }: LoaderArgs) => {
  const { sessionID } = params;
  invariant(sessionID, "sessionID is required");
  await requireUser(request);
  const details = await getRunSession({ id: sessionID });
  const runParams = getParameters(details.run);
  const runDevices = details.run.steps.map(({ action }) => action.device).filter((device, index, self) => {
    return self.findIndex((d) => d.id === device.id) === index;
  });
  return json({ session: details, runParams, runDevices });
};

type EventConstant = {
  event: RunSessionDetails["run"]["steps"][0]["action"]["event"]
  order: number
  constants: RunSessionDetails["run"]["steps"][0]["action"]["constants"]
  deviceID: string
}

function getParameters(run: RunSessionDetails["run"]): Record<string, Parameter[]> {
  // make all EventConstants from the run
  const eventConstants: EventConstant[] = run.steps.map(({ action, order }) => {
    return {
      order,
      deviceID: action.device.id,
      event: action.event!,
      constants: action.constants
    };
  });
  const params: Record<string, Parameter[]> = {};
  eventConstants.forEach(({ order, event, constants, deviceID }) => {
    event.fields.forEach((field) => {
      const constant = constants.find((constant) => constant.field.id === field.id);
      if (constant) return;
      if (!params[deviceID]) params[deviceID] = [];
      params[deviceID].push({
        order,
        eventID: event.id,
        deviceID: deviceID,
        fieldID: field.id,
        fieldName: field.name,
        fieldType: field.type
      });
    });
  });
  return params;
}

type DeviceCardProps = {
  device: RunSessionDetails["run"]["steps"][0]["action"]["device"]
  net: PetriNet
}

export function DeviceCard({ device, net }: DeviceCardProps) {
  const marking = useContextSelector(RunSessionContext, (context) => context?.session.markings[device.id]);
  return (
    <div
      className={"flex flex-col md:w-1/3 sm:w-1/2 lg:w-1/6 space-y-2 border-md rounded-lg p-2 border-2 border-slate-800 dark:border-slate-400"}>
      <h1 className={"text-lg font-semibold"}>{device.name}</h1>
      <div>
        <h2 className={"text-md font-semibold"}>Marking</h2>
        <div className={"flex flex-row space-x-2"}>
          {marking && Object.entries(marking).map(([placeID, tokens]) => {
              return (
                <div key={placeID} className={"flex flex-col space-y-2"}>
                  <h3 className={"text-sm font-semibold"}>{net.placeName(placeID)}</h3>
                  <div className={"flex flex-row space-x-2"}>
                    {tokens > 0 && Array.from(Array(tokens).keys()).map((token) => {
                      return (
                        <div
                          key={token}
                          className={"w-4 h-4 rounded-full bg-green-500"}
                        />
                      );
                    })}
                  </div>
                </div>
              );
            }
          )}
        </div>
      </div>
    </div>
  );
}

type ControlButtonProps = {
  submit?: boolean
  disabled: boolean
  color: "red" | "green" | "blue" | "yellow"
  children: ReactNode
}


export function ControlButton({ submit, disabled, color, children }: ControlButtonProps) {
  const setColor = (color: "red" | "green" | "blue" | "yellow") => {
    switch (color) {
      case "red":
        return "bg-rose-700 hover:bg-rose-600";
      case "green":
        return "bg-green-700 hover:bg-green-600";
      case "blue":
        return "bg-blue-700 hover:bg-blue-600";
      case "yellow":
        return "bg-yellow-700 hover:bg-yellow-600";
    }
  };
  return (
    <button
      type={`${submit ? "submit" : "button"}`}
      className={`rounded-full px-2 py-1 text-white ${setColor(color)}`}
      disabled={disabled}
    >
      {children}
    </button>
  );

}

export default function PlaySequence() {
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);
  const { session, runParams, runDevices } = useLoaderData<typeof loader>();
  const dispatch = useContextSelector(RunSessionContext, (context) => context?.dispatch);
  const deviceMarkings = useQuery(DeviceMarkingsDocument, {
    variables: {
      input: {
        instances: session.instances.map(({ id, device }) => {
          return {
            instanceID: id,
            deviceID: device.id
          };
        })
      }
    },
    pollInterval: 1000,
    onCompleted: (data) => {
      if (!dispatch) return;
      const mr: Record<string, Marking> = {};
      data.deviceMarkings.forEach((deviceMarking) => {
          mr[deviceMarking.deviceID] = deviceMarking.marking as Record<string, number>;
        }
      );
      dispatch(
        {
          type: RunSessionActionType.MarkingUpdated,
          payload: mr
        }
      );
      console.log(mr);
    }
  });

  const paramRecord = useRef<Record<string, any>>({});

  async function handleStart(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    formData.set("parameters", JSON.stringify(paramRecord.current));

    const submission = parse(formData, {
      schema: z.object({
        sequenceID: z.string(),
        userID: z.string(),
        parameters: z.preprocess((s) => JSON.parse(s as string), z.record(z.any()).optional()),
        instances: z.preprocess((instStrings) => {
          const instanceStrings = instStrings as string[];
          return instanceStrings.map((instString) => {
            return JSON.parse(instString.toString()) as DeviceInstanceInput;
          });
        }, z.array(z.object({
          deviceID: z.string(),
          instanceID: z.string()
        })))
      })
    });
    if (!submission.value || submission.intent !== "submit") {
      console.log(submission.error);
      return;
    }
    console.log(submission.value);

  }

  return (
    <Suspense fallback={<div>Loading...</div>}>
      {session && deviceMarkings.data &&
        <RunSessionProvider session={session} devices={deviceMarkings.data.deviceMarkings}
                            parameters={runParams}>
          <div className={"flex flex-col h-screen w-full items-center justify-items-center"}>
            {session && runParams && runDevices && petriNet &&
              runDevices.map((device) => {
                return (

                  <DeviceCard
                    key={device.id}
                    device={device}
                    net={petriNet.net}
                  />
                );
              })
            }
            <div className={"flex flex-col w-full space-y-2"}>
              <div className={"flex flex-row space-x-2 justify-end"}>
                <Form onSubmit={handleStart}>
                  <ControlButton
                    submit
                    disabled={false}
                    color={"green"}
                    aria-label={"Start the sequence"}
                  >
                    <PlayIcon className={"h-5 w-5"} />
                  </ControlButton>
                </Form>
                <Player devices={runDevices} />
              </div>
            </div>
          </div>
        </RunSessionProvider>
      }
    </Suspense>
  )
    ;
};