import { attribute as _, Digraph, Subgraph, Node, Edge, toDot } from "ts-graphviz";
import { instance } from "@viz-js/viz";
import { useRef } from "react";

const G = new Digraph();
const A = new Subgraph("A");
const node1 = new Node("node1", {
  [_.color]: "red"
});
const node2 = new Node("node2", {
  [_.color]: "blue"
});
const edge = new Edge([node1, node2], {
  [_.label]: "Edge Label",
  [_.color]: "pink"
});
G.addSubgraph(A);
A.addNode(node1);
A.addNode(node2);
A.addEdge(edge);

const dot = toDot(G);
export const Graph = () => {
  const graph = useRef<HTMLDivElement>(null);
  instance().then((viz) => {
    const element = viz.renderSVGElement(dot);
    graph.current?.appendChild(element);
  });
  return (
    <div ref={graph} />
  );
}

export default Graph;