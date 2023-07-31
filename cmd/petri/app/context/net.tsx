import { createContext, useReducer } from "react";
import type { ReactNode, Dispatch, Reducer } from "react";
import type { Net, Arc, Place, Transition } from "@prisma/client";
import type { FormInput } from "~/lib/components/formInput";

export type NetState = {
  id: Net["id"];
  places: Pick<Place, "id" | "name">[];
  transitions: Pick<Transition, "id" | "name">[];
  arcs: Pick<Arc, "id" | "placeID" | "transitionID" | "fromPlace">[];
}

export const NetContext = createContext<NetState>({ id: "", places: [], transitions: [], arcs: [] });
export const NetDispatchContext = createContext<Dispatch<NetAction> | null>(null);

type NetProviderProps = {
  value: NetState;
  children: ReactNode;
}

export function NetProvider(props: NetProviderProps) {
  const [net, dispatch] = useReducer<Reducer<NetState, NetAction>>(netReducer, props.value);
  return (
    <NetContext.Provider value={net}>
      <NetDispatchContext.Provider value={dispatch}>
        {props.children}
      </NetDispatchContext.Provider>
    </NetContext.Provider>
  );
}

export enum NetActionKind {
  AddPlace = "add-place",
  AddTransition = "add-transition",
  AddArc = "add-arc",
  DeletePlace = "delete-place",
  DeleteTransition = "delete-transition",
  DeleteArc = "delete-arc",
  UpdatePlace = "update-place",
  UpdateTransition = "update-transition",
}

interface NetAction {
  type: NetActionKind;
  payload: FormInput;
}

function netReducer(net: NetState, action: NetAction): NetState {
  switch (action.type) {
    case NetActionKind.AddPlace: {
      const p = action.payload as Place;
      if (net.places.find((place) => place.id === p.id)) {
        break;
      }
      net.places.push(p);
      break;
    }
    case NetActionKind.AddTransition: {
      const t = action.payload as Transition;
      if (net.transitions.find((transition) => transition.id === t.id)) {
        break;
      }
      net.transitions.push(t);
      break;
    }
    case NetActionKind.AddArc: {
      const t = action.payload as Arc;
      if (net.arcs.find((arc) => arc.id === t.id)) {
        break;
      }
      net.arcs.push(t);
      break;
    }
    case NetActionKind.DeletePlace: {
      const { id } = action.payload as { id: string };
      return {
        ...net,
        places: net.places.filter((place) => place.id !== id)
      };
    }
    case NetActionKind.DeleteTransition: {
      const { id } = action.payload as { id: string };
      return {
        ...net,
        transitions: net.transitions.filter((transition) => transition.id !== id)
      };
    }
    case NetActionKind.DeleteArc: {
      const { id } = action.payload as { id: string };
      return {
        ...net,
        arcs: net.arcs.filter((arc) => arc.id !== id)
      };
    }
    case NetActionKind.UpdatePlace: {
      const p = action.payload as Place;
      return {
        ...net,
        places: net.places.map((place) => {
          if (place.id === p.id) {
            return {
              ...place,
              ...action.payload
            };
          }
          return place;
        })
      };
    }
    case NetActionKind.UpdateTransition: {
      const t = action.payload as Transition;
      return {
        ...net,
        transitions: net.transitions.map((transition) => {
          if (transition.id === t.id) {
            return {
              ...transition,
              ...action.payload
            };
          }
          return transition;
        })
      };
    }
    default:
      break;

  }
  return net;
}

