import * as React from "react";
import {  useRef } from "react";
import { render } from "~/lib/util/mermaid";
import type { MermaidConfig } from "mermaid";
import type { NetState } from "~/context/net";

type DiagramProps = {
  net: NetState;
  clicked: (id: string, kind: "place" | "transition") => void
}

export const MermaidDiagram = (props: DiagramProps) => {
  const container = useRef<HTMLDivElement>(null);

  const places = props.net.places.map((place) => {
    if (place.name === "") {
      return `p_${place.id}`;
    }
    return `p_${place.id}((${place.name}))\n  click p_${place.id} clicked`;
  });
  const transitions = props.net.transitions.map((transition) => {
    if (transition.name === "") {
      return `t_${transition.id}`;
    }
    return `t_${transition.id}[${transition.name}] \n  click t_${transition.id} clicked`;
  });
  const arcs = props.net.arcs.map((arc) => {
    if (arc.fromPlace) {
      return `p_${arc.placeID} --> t_${arc.transitionID}`;

    } else {
      return `t_${arc.transitionID} --> p_${arc.placeID}`;
    }
  });
  const out = `graph TD
  ${places.join("\n  ")}
  ${transitions.join("\n  ")}
  ${arcs.join("\n  ")}
  `;

  const handleClicked = (id: string) => {
    let [k, uuid] = id.split("_");
    let kind: "place" | "transition" = "transition";
    if (k === "p") {
      kind = "place";
    }
    props.clicked(uuid, kind);
  };
  if (container.current !== null) {
    // @ts-ignore
    window.clicked = (e: string) => {
      handleClicked(e);
    };
    render(
      { theme: "default", securityLevel: "loose" } as MermaidConfig,
      out,
      "graph-div").then(({ svg, bindFunctions }) => {
      container.current!.innerHTML = svg;
      const gd = document.getElementById("graph-div");
      if (gd) {
        if (bindFunctions) {
          bindFunctions(gd!);
        }
      }
    });
  }

  return (
    <div className={"flex flex-row w-full place-items-center"}>
      <div className={"flex flex-row w-full"} ref={container} />
    </div>
  );
};