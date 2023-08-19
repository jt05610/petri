import type { PetriNet } from "~/util/petrinet";
import { Suspense, useEffect, useState } from "react";
import { Graphviz } from "graphviz-react";
import type { ColorProfile } from "~/lib/components/markedNet";
import { colorProfiles } from "~/lib/components/markedNet";

type LabeledNetProps = {
  net: PetriNet,
  colorProfile?: "default" | ColorProfile,
  placeSize?: number,
}

export function LabeledNet(props: LabeledNetProps) {
  const [graph, setGraph] = useState<string>("");
  useEffect(() => {
    const graph = props.net.toGraphViz(
      props.placeSize || 1,
      colorProfiles["default"],
    );
    setGraph(graph);
  }, [props]);

  return (
    <div className="flex flex-col items-center border dark:border-slate-200 border-md rounded-lg m-2">
      <Suspense fallback={<div>Loading...</div>}>
        {graph && (
          <Graphviz
            dot={graph}
            options={{ zoom: true, height: 720, width: 1280 }} />
        )}
      </Suspense>
    </div>
  );
}
