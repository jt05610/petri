import type { PetriNet } from "~/util/petrinet";
import { Suspense, useEffect, useState } from "react";
import { Graphviz } from "graphviz-react";

type LabeledNetProps = {
  net: PetriNet,
}

export function LabeledNet(props: LabeledNetProps) {
  const [graph, setGraph] = useState<string>("");
  useEffect(() => {
    const graph = props.net.toGraphViz();
    setGraph(graph);
  }, [props]);
  return (
    <Suspense fallback={<div>Loading...</div>}>
      {graph && <Graphviz dot={graph} options={{ zoom: true }} />}
    </Suspense>
  );
}