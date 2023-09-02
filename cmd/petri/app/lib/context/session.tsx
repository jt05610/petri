import { createContext } from "use-context-selector";
import type { Dispatch, ReactNode, Reducer } from "react";
import { useReducer } from "react";
import type { DataListItem, RunSessionDetails } from "~/models/net.run.session.server";
import type { Datum } from "@prisma/client";
import type { DeviceMarking, StartSessionInput } from "~/models/__generated__/graphql";
import type { DatumInput, Parameter } from "~/models/net.run.session.data.server";
import type { Marking } from "~/util/petrinet";
import { SessionState } from "@prisma/client";

export enum RunSessionActionType {
  Start = "start",
  Stop = "stop",
  Pause = "pause",
  Resume = "resume",
  ChangeParameter = "changeParameter",
  Step = "step",
  DataReceived = "dataReceived",
  MarkingUpdated = "markingUpdated"
}


export type ParameterWithValue = {
  parameter: Parameter
  value: string | number | boolean
}


type RunSessionAction = {
  type: RunSessionActionType
  payload?: StartSessionInput | string | ParameterWithValue | DatumInput | Record<string, Marking>
}

type RunSessionState = {
  parameters?: Record<string, Record<number, Record<string, ParameterWithValue>>>
  status: SessionState
  startedAt: Date | string | null
  stoppedAt: Date | string | null
  pausedAt: Date | string | null
  markings: Record<string, Marking>
  currentStep: number
  data: DataListItem[]
}

type RunSessionContext = {
  session: RunSessionState
  dispatch: Dispatch<RunSessionAction>
}

function sessionReducer(state: RunSessionState, action: RunSessionAction): RunSessionState {
  switch (action.type) {
    case RunSessionActionType.Start:
      return {
        ...state,
        status: SessionState.RUNNING,
        startedAt: new Date()
      };
    case RunSessionActionType.Stop:
      return {
        ...state,
        status: SessionState.STOPPED,
        stoppedAt: new Date()
      };
    case RunSessionActionType.Pause:
      return {
        ...state,
        status: SessionState.PAUSED,
        pausedAt: new Date()
      };
    case RunSessionActionType.Resume:
      return {
        ...state,
        status: SessionState.RUNNING,
        pausedAt: null
      };
    case RunSessionActionType.Step:
      return {
        ...state,
        currentStep: state.currentStep + 1
      };
    case RunSessionActionType.MarkingUpdated:
      const { marking } = action.payload as Record<string, Marking>;
      return {
        ...state,
        markings: {
          ...state.markings,
          marking
        }
      };
    case RunSessionActionType.DataReceived:
      const datum = action.payload as Datum;
      return {
        ...state,
        data: [...state.data, datum]
      };
    case RunSessionActionType.ChangeParameter:
      if (!state.parameters) return state;
      const { parameter, value } = action.payload as ParameterWithValue;
      const { deviceID, fieldID, order } = parameter;
      return {
        ...state,
        parameters: {
          ...state.parameters,
          [deviceID]: {
            ...state.parameters[deviceID],
            [order]: {
              ...state.parameters[deviceID]?.[order],
              [fieldID]: {
                parameter,
                value
              }
            }
          }
        }
      };

    default:
      return state;
  }
}

export const RunSessionContext = createContext<RunSessionContext | null>(null);

type RunSessionProviderProps = {
  session: RunSessionDetails
  devices: DeviceMarking[]
  children: ReactNode
  parameters?: Record<string, Record<number, Record<string, ParameterWithValue>>>
}

function toMarkingRecord(devices: DeviceMarking[]): Record<string, Marking> {
  const mr: Record<string, Marking> = {};
  devices.forEach((deviceMarking) => {
      mr[deviceMarking.deviceID] = deviceMarking.marking as Record<string, number>;
    }
  );
  return mr;
}


export function RunSessionProvider({ session, devices, children, parameters }: RunSessionProviderProps) {
  const [runSession, dispatch] = useReducer<Reducer<RunSessionState, RunSessionAction>>(sessionReducer, {
    parameters: parameters,
    status: session.state,
    startedAt: session.startedAt,
    stoppedAt: session.stoppedAt,
    pausedAt: session.pausedAt && session.pausedAt[session.pausedAt.length - 1],
    markings: toMarkingRecord(devices),
    currentStep: 0,
    data: session.data
  });
  console.log(runSession.parameters);
  return (
    <RunSessionContext.Provider value={{ session: runSession, dispatch }}>
      {children}
    </RunSessionContext.Provider>
  );
}