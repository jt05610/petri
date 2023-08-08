import { Graphviz } from "graphviz-react";
import { Suspense, useEffect, useState } from "react";
import colors from "tailwindcss/colors";
import type { Marking, PetriNet } from "~/util/petrinet";

type MarkedNetProps = {
  colorProfile?: "default" | ColorProfile,
  net: PetriNet,
  marking: Marking,
  placeSize?: number,
}

export const colorProfiles: {
  [key: string]: ColorProfile
} = {
  "default": {
    bg: colors.slate[900],
    text: colors.slate[100],
    border: colors.gray[200],
    placeColor: colors.gray[400],
    placeTokenColor: colors.amber[600],
    transitionColor: colors.gray[600],
    enabledTransition: colors.rose[500],
    enabledArc: colors.red[900]
  }
};

export type ColorProfile = {
  bg: string
  text: string
  border: string
  transitionColor: string
  placeColor: string
  placeTokenColor: string
  enabledTransition: string
  enabledArc: string
}

export function MarkedNet(props: MarkedNetProps) {
  const [graph, setGraph] = useState<string>("");

  useEffect(() => {
    setGraph(props.net.toGraphVizWithMarking(
      props.placeSize || 1,
      colorProfiles["default"],
      props.marking
    ) || "");
  }, [props.net, props.marking, props.placeSize, props.colorProfile]);

  return (
    <div className="flex flex-col items-center border dark:border-slate-200 border-md rounded-lg m-2">
      <Suspense fallback={<div>Loading...</div>}>
        {graph && (
          <Graphviz
            dot={graph}
            options={{ zoom: false, height: 720, width: 1280 }} />
        )}
      </Suspense>
    </div>
  );
}