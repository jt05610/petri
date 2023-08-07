import type { PetriNet } from "~/util/petrinet";
import { Graphviz } from "graphviz-react";
import { Suspense, useEffect, useState } from "react";

type MarkedNetProps = {
  marking: { [key: string]: number },
  net: PetriNet,
}

export function MarkedNet(props: MarkedNetProps) {
  const [graph, setGraph] = useState<string>("");
  useEffect(() => {
    const graph = props.net.toGraphVizWithMarking(props.marking);
    setGraph(graph);
  }, [props.net, props.marking]);
  return (
    <Suspense fallback={<div>Loading...</div>}>
      {graph &&
        <Graphviz
          dot={props.net.toGraphVizWithMarking(props.marking)}
          options={{ zoom: false, fit: true, height: 720, width: 1280 }} />
      }
    </Suspense>
  )
    ;
}