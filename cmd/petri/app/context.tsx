import type { ReactNode } from "react";
import { createContext, useContextSelector } from "use-context-selector";
import type { Socket } from "socket.io-client";
import { PetriNet } from "~/util/petrinet";
import { useState } from "react";
import type { NetDetails } from "~/models/net.server";
import type { Place } from "@prisma/client";

type ProviderProps = {
  socket: Socket | undefined;
  children: ReactNode;
};

const socketContext = createContext<Socket | undefined>(undefined);

export function useSocket() {
  return useContextSelector(socketContext, (socket) => socket);
}

export function SocketProvider({ socket, children }: ProviderProps) {
  return <socketContext.Provider value={socket}>{children}</socketContext.Provider>;
}

export const PetriNetContext = createContext<({
  petriNet: PetriNet;
  marking: { [key: string]: number };
  setMarking: (marking: { [key: string]: number }) => void;
}) | null>(null);

type PetriNetProviderProps = {
  net: NetDetails;
  children: ReactNode;
}
const netMarking = (places: Pick<Place, "id">[], initial: number[]) => {
  return places.map((p, i) => [p.id, initial[i]]).reduce((acc, [id, tokens]) => ({ ...acc, [id]: tokens }), {});
};

export function PetriNetProvider({ net, children }: PetriNetProviderProps) {
  const [petriNet] = useState<PetriNet>(new PetriNet(net));
  const [marking, setMarking] = useState<{ [key: string]: number }>(netMarking(net.places, net.initialMarking));
  return (
    <PetriNetContext.Provider value={{ petriNet, marking, setMarking }}>
      {children}
    </PetriNetContext.Provider>
  );
}
