import type { ReactNode } from "react";
import { createContext, useContextSelector } from "use-context-selector";
import type { NetDetails } from "~/models/net.server";

interface NetContextType {
  net: NetDetails;
}

export const NetContext = createContext<NetContextType | null>(null);

export const NetProvider = (props: { children: ReactNode, net: NetDetails }) => {
  return (
    <NetContext.Provider value={{ net: props.net }}>
      {props.children}
    </NetContext.Provider>
  );
};

export const useNet = () => {
  const net = useContextSelector(NetContext, (state) => state ? state.net : null);
  if (!net) {
    throw new Error("useNet must be used within a NetProvider");
  }
  return net;
};
