import { PetriNet } from "~/util/petrinet";

test("PetriNet", () => {
  const net = new PetriNet({
    id: "1",
    name: "test",
    description: "test",
    device: {
      instances: [
        {
          id: "1",
          addr: "1",
          name: "test"
        }
      ]
    },
    initialMarking: [1, 0],
    places: [
      {
        id: "1",
        name: "A",
        bound: 1
      },
      {
        id: "2",
        name: "B",
        bound: 1
      }
    ],
    transitions: [
      {
        id: "3",
        name: "A Opened"
      },
      {
        id: "4",
        name: "B Opened",
        events: [
          { id: "1", name: "open B", fields: [] }
        ]
      }
    ],
    arcs: [
      {
        placeID: "1",
        transitionID: "3",
        fromPlace: false
      },
      {
        placeID: "2",
        transitionID: "3",
        fromPlace: true
      },
      {
        placeID: "2",
        transitionID: "4",
        fromPlace: false
      },
      {
        placeID: "1",
        transitionID: "4",
        fromPlace: true
      }
    ],
    placeInterfaces: [],
    transitionInterfaces: [],
    children: []
  });
  expect(net).toBeInstanceOf(PetriNet);

  expect(net.nodes).toHaveLength(4);
  expect(net.arcs).toHaveLength(4);
  expect(net.enabledTransitions).toHaveLength(1);
});

test("PetriNet#graphviz", () => {
  const net = new PetriNet({
    id: "1",
    name: "test",
    description: "test",
    placeInterfaces: [],
    transitionInterfaces: [],
    children: [],
    device: {
      instances: [
        {
          id: "1",
          addr: "1",
          name: "test"
        }
      ]
    },
    initialMarking: [1, 0],
    places: [
      {
        id: "1",
        name: "A",
        bound: 1
      },
      {
        id: "2",
        name: "B",
        bound: 1
      }
    ],
    transitions: [
      {
        id: "1",
        name: "A Opened"
      },
      {
        id: "2",
        name: "B Opened",
        events: [
          { id: "1", name: "open B", fields: [] }
        ]
      }
    ],
    arcs: [
      {
        placeID: "1",
        transitionID: "1",
        fromPlace: false
      },
      {
        placeID: "2",
        transitionID: "1",
        fromPlace: true
      },
      {
        placeID: "2",
        transitionID: "2",
        fromPlace: false
      },
      {
        placeID: "1",
        transitionID: "2",
        fromPlace: true
      }
    ]
  });
  expect(net.toGraphViz({ rankdir: "LR" })).toBe(`digraph {
    rankdir=LR;
    node [shape=circle];
    "p1" [label="A"];
    "p2" [label="B"];
    "t1" [label="A Opened" shape=box];
    "t2" [label="B Opened" shape=box];
    "t1" -> "p1";
    "p2" -> "t1";
    "t2" -> "p2";
    "p1" -> "t2";
}`);

  expect(net.toGraphVizWithMarking({ 1: 1, 2: 0 }, { rankdir: "LR" })).toBe(`digraph {
    rankdir=LR;
    node [shape=circle];
    "p1" [label="A" style=filled fillcolor=green];
    "p2" [label="B"];
    "t1" [label="A Opened" shape=box];
    "t2" [label="B Opened" shape=box];
    "t1" -> "p1";
    "p2" -> "t1";
    "t2" -> "p2";
    "p1" -> "t2";
}`);
});