import type { Dispatch, ReactNode, Reducer } from "react";
import { createContext } from "use-context-selector";
import type { Socket } from "socket.io-client";
import { PetriNet } from "~/util/petrinet";
import { useReducer, useState } from "react";
import type { NetDetailsWithChildren } from "~/models/net.server";
import type {
  ActionInputDisplay,
  RunInputDisplay
} from "~/models/net.run.server";
import type { Place, PrismaClient } from "@prisma/client";
import type { DataListItem } from "~/models/net.run.session.data.server";
import type { RunSessionDetails } from "~/models/net.run.session.server";

export const PrismaContext = createContext<PrismaClient | undefined>(undefined);

export type PrismaProviderProps = {
  prisma: PrismaClient;
  children: ReactNode;
}

export function PrismaProvider({ prisma, children }: PrismaProviderProps) {
  return <PrismaContext.Provider value={prisma}>{children}</PrismaContext.Provider>;
}

type ProviderProps = {
  socket: Socket | undefined;
  children: ReactNode;
};

export const SocketContext = createContext<Socket | undefined>(undefined);

export function SocketProvider({ socket, children }: ProviderProps) {
  return <SocketContext.Provider value={socket}>{children}</SocketContext.Provider>;
}

export const PetriNetContext = createContext<({
  petriNet: PetriNet;
  marking: {
    [key: string]: number
  };
  setMarking: (marking: {
    [key: string]: number
  }) => void;
}) | null>(null);

type PetriNetProviderProps = {
  net: NetDetailsWithChildren;
  children: ReactNode;
}
const netMarking = (places: Pick<Place, "id">[], initial: number[]) => {
  return places.map((p, i) => [p.id, initial[i]]).reduce((acc, [id, tokens]) => ({ ...acc, [id]: tokens }), {});
};

export function PetriNetProvider({ net, children }: PetriNetProviderProps) {

  const [petriNet] = useState<PetriNet>(net.devices ? new PetriNet(net) : new PetriNet(net).combinedNet);
  const [marking, setMarking] = useState<{
    [key: string]: number
  }>(netMarking(petriNet.net.places, petriNet.net.initialMarking));
  return (
    <PetriNetContext.Provider value={{ petriNet, marking, setMarking }}>
      {children}
    </PetriNetContext.Provider>
  );
}

export const RecordRunContext = createContext<({
  run: RunInputDisplay;
  dispatch: Dispatch<RunAction>;
}) | null>(null);

type ActionAddedPayload = ActionInputDisplay;


type ActionRemovedPayload = {
  index: number;
}

type ConstantDeletedPayload = {
  fieldID: string;
  actionIndex: number;
}

type RunActionPayload = ActionAddedPayload | ActionRemovedPayload | ConstantDeletedPayload;

export enum RunActionType {
  ActionAdded = "actionAdded",
  ActionRemoved = "actionRemoved",
  ConstantDeleted = "removeConstant",
}

type RunAction = {
  type: RunActionType;
  payload: RunActionPayload;
}

const defaultInitial: RunInputDisplay = {
  deviceNames: [],
  description: "",
  name: "",
  netID: "",
  actions: []
};


type RecordRunProviderProps = {
  initialRun?: RunInputDisplay;
  children: ReactNode;
}

export const addDeviceNames = (run: RunInputDisplay, pNet: PetriNet): RunInputDisplay => {
  return {
    ...run,
    deviceNames: pNet.devices.map((d) => d.name)
  };
};

export function RecordRunProvider({ initialRun, children }: RecordRunProviderProps) {
  const [run, dispatch] = useReducer<Reducer<RunInputDisplay, RunAction>>(runReducer, initialRun || defaultInitial);
  return (
    <RecordRunContext.Provider value={{ run, dispatch }}>
      {children}
    </RecordRunContext.Provider>
  );
}

function runReducer(state: RunInputDisplay, action: RunAction): RunInputDisplay {
  switch (action.type) {
    case RunActionType.ActionAdded: {
      const payload = action.payload as ActionAddedPayload;
      const actions = [...state.actions];
      actions.push(payload);
      return { ...state, actions };
    }
    case RunActionType.ActionRemoved: {
      const { index } = action.payload as ActionRemovedPayload;
      const actions = [...state.actions];
      actions.splice(index, 1);
      return { ...state, actions };
    }
    case RunActionType.ConstantDeleted: {
      const { fieldID, actionIndex } = action.payload as ConstantDeletedPayload;
      const actions = [...state.actions];
      const runAction = actions[actionIndex];
      const constants = runAction.constants.filter((c) => c.fieldID !== fieldID);
      actions[actionIndex] = { ...runAction, constants };
      return { ...state, actions };
    }
    default:
      throw Error("Invalid action type");
  }
}

export enum SessionActionType {
  StartSession = "startSession",
  StopSession = "stopSession",
  ActionStarted = "actionStarted",
  ActionCompleted = "actionCompleted",
}

type ActionStartedPayload = {}

type ActionCompletedPayload = DataListItem

type SessionActionPayload = ActionStartedPayload | ActionCompletedPayload;

type SessionAction = {
  type: SessionActionType;
  payload: SessionActionPayload;
}

type LiveSession = RunSessionDetails & {
  remainingActions: RunSessionDetails["run"]["steps"];
  activeIndex: number;
  activeAction: RunSessionDetails["run"]["steps"][number] | undefined;
  data: DataListItem[];
}

type SessionContext = {
  session: LiveSession
  dispatch: Dispatch<SessionAction>;
}

export const RunSessionContext = createContext<SessionContext | null>(null);

type SessionProviderProps = {
  sessionDetails: RunSessionDetails;
  children: ReactNode;
}

function createLiveSession(session: RunSessionDetails): LiveSession {
  return {
    ...session,
    remainingActions: session.run.steps,
    activeAction: session.run.steps[0],
    activeIndex: 0,
    data: [] as DataListItem[]
  };
}

export function SessionProvider({ sessionDetails, children }: SessionProviderProps) {
  const [session, dispatch] = useReducer<Reducer<LiveSession, SessionAction>>(sessionReducer, createLiveSession(sessionDetails));

  return (
    <RunSessionContext.Provider value={{ session, dispatch }}>
      {children}
    </RunSessionContext.Provider>
  );
}

function sessionReducer(state: LiveSession, action: SessionAction): LiveSession {
  switch (action.type) {
    case SessionActionType.StartSession: {
      const newState = state;
      newState.activeAction = state.remainingActions[0];
      newState.activeIndex = 0;
      return newState;
    }
    case SessionActionType.ActionStarted: {
      const remainingActions = [...state.remainingActions];
      const activeAction = remainingActions.shift();
      return {
        ...state,
        remainingActions,
        activeAction: activeAction,
        activeIndex: state.activeIndex + 1
      };
    }
    case SessionActionType.ActionCompleted: {
      const payload = action.payload as ActionCompletedPayload;
      const data = [...state.data];
      data.push(payload);
      return {
        ...state,
        data,
        activeAction: undefined
      };
    }
    case SessionActionType.StopSession: {
      return createLiveSession(state);
    }
    default:
      throw Error("Invalid action type");
  }
}

