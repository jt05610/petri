import type { Dispatch, ReactNode, Reducer } from "react";
import { useReducer } from "react";
import { createContext, useContextSelector } from "use-context-selector";
import cloneDeep from "lodash/cloneDeep";

interface ParserState {
  scope: Map<string, any>;
  error?: Error;
}

interface ParserAction {
  handle: (state: ParserState) => ParserState;
}

export const addToScope = (key: string, value: string): ParserAction => {
  return {
    handle: (state: ParserState) => {
      try {
        const scope = cloneDeep(state.scope);
        scope.set(key, value);
        return {
          ...state,
          scope,
          error: undefined
        };
      } catch (e) {
        return {
          ...state,
          error: e as Error
        };
      }
    }
  };
};

export const removeFromScope = (key: string): ParserAction => {
  return {
    handle: (state: ParserState) => {
      return {
        ...state,
        scope: new Map([...state.scope].filter(([k]) => k !== key))
      };
    }
  };
};

interface ParserContextType {
  state: ParserState;
  dispatch: Dispatch<ParserAction>;
}

export const ParserContext = createContext<ParserContextType | null>(null);

const parserReducer = (state: ParserState, action: ParserAction): ParserState => {
  return action.handle(state);
};

type ParserProviderProps = {
  children: ReactNode;
  scope: Map<string, any>;
}

export const ParserProvider = ({ children, scope }: ParserProviderProps) => {
  const [state, dispatch] = useReducer<Reducer<ParserState, ParserAction>>(parserReducer, {
    scope: scope
  });
  return (
    <ParserContext.Provider value={{ state, dispatch }}>
      {children}
    </ParserContext.Provider>
  );

};


export const useParserContext = () => {
  const context = useContextSelector(ParserContext, (context) => context);
  if (!context) {
    throw new Error("useParserContext must be used within a ParserProvider");
  }
  return context;
};