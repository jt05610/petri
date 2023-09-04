import type { NetListItem } from "~/models/net.server";
import { createContext } from "use-context-selector";
import type { ReactNode } from "react";


export const NetsContext = createContext<NetListItem[] | null>(null);

type NetsProviderProps = {
  nets: NetListItem[],
  children: ReactNode
}

export function NetsProvider({ nets, children }: NetsProviderProps) {
  return (
    <NetsContext.Provider value={nets}>
      {children}
    </NetsContext.Provider>
  )
    ;
}