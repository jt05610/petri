import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import type { RunSessionDetails } from "~/models/net.run.session.server";
import { getRunSession } from "~/models/net.run.session.server";
import invariant from "tiny-invariant";
import { requireUser } from "~/session.server";
import { useLoaderData } from "@remix-run/react";
import Player from "~/lib/components/player";
import type { StartSessionInput } from "~/models/__generated__/graphql";
import {
  DeviceMarkingsDocument, NewEventsDocument,
  PauseSessionDocument, ResumeSessionDocument,
  StartDocument, StopSessionDocument
} from "~/models/__generated__/graphql";
import { useRef, type ReactNode, Suspense } from "react";
import type { ParameterWithValue } from "~/lib/context/session";
import { RunSessionActionType, RunSessionContext, RunSessionProvider } from "~/lib/context/session";
import { useMutation, useQuery } from "@apollo/client";
import { useContextSelector } from "use-context-selector";
import type { Marking, PetriNet } from "~/util/petrinet";
import { PetriNetContext } from "~/lib/context/petrinet";
import { PlayIcon } from "@heroicons/react/24/solid";
import { SessionState } from "@prisma/client";
import { PauseIcon, StopIcon } from "@heroicons/react/24/outline";
import { getParameters, toParameterRecord } from "~/util/parameters";

export const loader = async ({ params, request }: LoaderArgs) => {
  const { sessionID } = params;
  invariant(sessionID, "sessionID is required");
  await requireUser(request);
  const details = await getRunSession({ id: sessionID });
  const parameters = getParameters(details.run);
  const runDevices = details.run.steps.map(({ action }) => action.device).filter((device, index, self) => {
    return self.findIndex((d) => d.id === device.id) === index;
  });
  console.log(runDevices);
  return json({ session: details, runParams: toParameterRecord(parameters), runDevices });
};

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
  onClick: () => void
}


export function ControlButton({ submit, disabled, color, children, onClick }: ControlButtonProps) {
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
      className={`rounded-full px-2 py-1 text-white ${setColor(color)}`}
      disabled={disabled}
      onClick={onClick}
    >
      {children}
    </button>
  );

}

export default function PlaySequence() {
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);
  const { session, runParams, runDevices } = useLoaderData<typeof loader>();
  const activeSession = useContextSelector(RunSessionContext, (context) => context?.session);
  const paramRef = useRef<Record<string, Record<number, Record<string, ParameterWithValue>>>>(runParams);
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


  const [start] = useMutation(
    StartDocument, {});

  const [pause] = useMutation(
    PauseSessionDocument, {});

  const [stop] = useMutation(
    StopSessionDocument, {});

  useQuery(
    NewEventsDocument, {
      variables: {
        id: session.id
      },
      onCompleted: (data) => {
        if (data.newEvents.length === 0) return;
        if (!dispatch) return;
        dispatch(
          {
            type: RunSessionActionType.Step
          }
        );
      },
      pollInterval: 1000
    });

  const [resume] = useMutation(
    ResumeSessionDocument, {}
  );

  async function handlePause() {
    const res = pause({ variables: { input: session.id } }).then((res) => {
      return res.data?.pauseSession;
    });
    if (!res) return;
    if (!dispatch) return;
    dispatch(
      {
        type: RunSessionActionType.Pause
      });
  }

  async function handleResume() {
    const res = resume({ variables: { input: session.id } }).then((res) => {
      return res.data?.resumeSession;
    });
    if (!res) return;
    if (!dispatch) return;
    dispatch(
      {
        type: RunSessionActionType.Resume
      });
  }

  async function handleStop() {
    const res = stop({ variables: { input: session.id } }).then((res) => {
      return res.data?.stopSession;
    });
    if (!res) return;
    if (!dispatch) return;
    dispatch(
      {
        type: RunSessionActionType.Stop
      });
  }

  async function handleStart() {
    console.log("start");
    const submission: StartSessionInput = {
      sessionID: session.id,
      parameters: paramRef.current
    };
    console.log("submitting", submission);
    const ok = await start({ variables: { input: submission } }).then((res) => {
      return res.data?.startSession;
    });
    if (!ok) return;
  }


  return (
    <Suspense fallback={<div>Loading...</div>}>
      {session && deviceMarkings.data && runParams &&
        <RunSessionProvider session={session} devices={deviceMarkings.data.deviceMarkings}
                            parameters={runParams}>
          <div className={"flex flex-wrap h-screen w-full items-center justify-items-center"}>
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
                <ControlButton
                  disabled={activeSession?.status === SessionState.RUNNING}
                  color={"green"}
                  aria-label={"Start the sequence"}
                  onClick={handleStart}
                >
                  <PlayIcon className={"h-5 w-5"} />
                </ControlButton>
                {activeSession?.status === SessionState.RUNNING ? (
                  <ControlButton
                    disabled={activeSession?.status !== SessionState.RUNNING}
                    color={"yellow"}
                    aria-label={"Pause the sequence"}
                    onClick={handlePause}
                  >
                    <PauseIcon className={"h-5 w-5"} />
                  </ControlButton>
                ) : activeSession?.status === SessionState.PAUSED ? (
                  <ControlButton
                    disabled={activeSession?.status !== SessionState.PAUSED}
                    color={"blue"}
                    aria-label={"Resume the sequence"}
                    onClick={handleResume}
                  >
                    <PlayIcon className={"h-5 w-5"} />
                  </ControlButton>
                ) : null
                }
                <ControlButton
                  disabled={activeSession?.status !== SessionState.RUNNING}
                  color={"red"}
                  aria-label={"Stop the sequence"}
                  onClick={handleStop}
                >
                  <StopIcon className={"h-5 w-5"} />
                </ControlButton>
              </div>
              <Player devices={runDevices} paramRef={paramRef} />
            </div>
          </div>
        </RunSessionProvider>
      }
    </Suspense>
  )
    ;
};