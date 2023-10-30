import { Dispatch, ReactNode, Reducer, useReducer } from "react";
import { createContext } from "use-context-selector";

export function createDataContext<A, S>(initialState: S, reducer: Reducer<S, A>) {
  interface ContextProps {
    state: S;
    dispatch: Dispatch<A>;
  }

  const Context = createContext<ContextProps | undefined>(undefined);

  function Provider({ initial, children }: { initial?: S, children: ReactNode }) {
    const [state, dispatch] = useReducer(reducer, initial || initialState);

    return <Context.Provider value={{ state, dispatch }}>{children}</Context.Provider>;
  }

  return [Context, Provider] as const;
}