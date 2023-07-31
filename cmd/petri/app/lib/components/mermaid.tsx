import * as React from "react";
import { useEffect, useRef, useState } from "react";
import { render } from "~/lib/util/mermaid";
import type { MermaidConfig } from "mermaid";
import type { NetState } from "~/context/net";

type DiagramProps = {
  net: NetState
  setSelected: (update: { id: string, kind: string }) => void
}

export const MermaidDiagram = (props: DiagramProps) => {
  const container = useRef<HTMLDivElement>(null);
  const [error, setError] = useState(false);

  const places = props.net.places.map((place) => {
    return `${place.id}((${place.name}))\n  click ${place.id} clicked_p`;
  });
  const transitions = props.net.transitions.map((transition) => {
    return `${transition.id}[${transition.name}] \n  click ${transition.id} clicked_t`;
  });
  const arcs = props.net.arcs.map((arc) => {
    console.log(arc);
    if (arc.fromPlace) {
      return `${arc.placeID} --> ${arc.transitionID}`;

    } else {
      return `${arc.transitionID} --> ${arc.placeID}`;
    }
  });
  let out = `graph TD
  ${places.join("\n  ")}
  ${transitions.join("\n  ")}
  ${arcs.join("\n  ")}
  `;
  console.log(out);

  useEffect(() => {
    render(
      { theme: "default", securityLevel: "loose" } as MermaidConfig,
      out,
      "graph-div"
    ).then(({ svg, bindFunctions }) => {
      // @ts-ignore
      window.clicked_t = (id: string) => {
        props.setSelected({id, kind: "transition" });
      };
      // @ts-ignore
      window.clicked_p = (id: string) => {
        props.setSelected({id, kind: "place" });
      }
      if (!container.current) {
        setError(true);
        return;
      }
      container.current.innerHTML = svg;
      const gd = document.getElementById("graph-div");
      if (!gd) {
        setError(true);
        return;
      }
      if (bindFunctions) {
        bindFunctions(gd);
      }
    });
  });

  return (
    <div className={"flex flex-row w-full place-items-center"}>
      {error ? (
        <div id="errorMessage">
          {/* your error message */}
        </div>
      ) : <div className={"flex flex-row w-full"} ref={container} />}

    </div>
  );
};