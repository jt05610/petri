import { createContext } from "use-context-selector";
import type {
  RunDetails
} from "~/models/net.run.server";
import type { ReactNode } from "react";

type RunContextProps = {
  run: RunDetails
}

export const RunContext = createContext<RunContextProps | null>(null);

type RunProviderProps = {
  runDetails: RunDetails
  children: ReactNode
}

export const RunProvider = ({ runDetails, children }: RunProviderProps) => {
  return (
    <RunContext.Provider value={{ run: runDetails }}>
      {children}
    </RunContext.Provider>
  );
};