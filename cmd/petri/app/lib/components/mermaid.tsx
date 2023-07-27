import * as React from "react";
import { useRef, useState } from "react";
import { render } from "~/lib/util/mermaid";
import type { MermaidConfig } from "mermaid";

type Props = {
  code: string
};


export const MermaidDiagram = (props: Props) => {
  const container = useRef<HTMLDivElement>(null);
  const [error, setError] = useState(false);

  render(
    { theme: "default", securityLevel: "loose" } as MermaidConfig,
    props.code,
    "graph-div"
  ).then(({ svg, bindFunctions }) => {
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

  return (
    <>
      {error && (
        <div id="errorMessage">
          {/* your error message */}
        </div>
      )}
      <div ref={container} />
    </>
  );
};