import { expect, describe, it } from "vitest";
import { controller } from "~/models/controller";
import { builder } from "~/models/builder";
import type { Net } from "~/models/net";

const places = [
  {
    name: "p1",
    bound: 1,
    initialMarking: 1,
    marking: 1
  },
  {
    name: "p2",
    bound: 1,
    initialMarking: 0,
    marking: 0
  }
];

const transitions = [
  {
    name: "t1",
    events: []
  },
  {
    name: "t2",
    events: []
  }
];

export function makeNet(): Net.PetriNet {
  const net = builder.addTransition(
    builder.addPlace(
      builder.newNet("test"), ...places), ...transitions
  );
  const arcs = [
    {
      from: net.places[0],
      to: net.transitions[0]
    },
    {
      from: net.transitions[0],
      to: net.places[1]
    },
    {
      from: net.places[1],
      to: net.transitions[1]
    },
    {
      from: net.transitions[1],
      to: net.places[0]
    }
  ];
  return builder.addArc(net, ...arcs);
}

describe.concurrent("controller", async () => {
  it("should return available transitions", async () => {
    const net = makeNet();
    const availableTransitions = controller.availableTransitions(net);
    expect(availableTransitions).toHaveLength(1);
    expect(availableTransitions[0].name).toBe("t1");
    const newNet = controller.fireTransition(net, availableTransitions[0]);
    const availableTransitions2 = controller.availableTransitions(newNet);
    expect(availableTransitions2).toHaveLength(1);
    expect(availableTransitions2[0].name).toBe("t2");
  });
  it("should fire a transition", async () => {
    const net = makeNet();
    const availableTransitions = controller.availableTransitions(net);
    const newNet = controller.fireTransition(net, availableTransitions[0]);
    expect(newNet.places[0].marking).toBe(0);
    expect(newNet.places[1].marking).toBe(1);
  });
  it("should return inputs", async () => {
    const net = makeNet();
    const inputs = controller.inputs(net, net.transitions[0]);
    expect(inputs).toHaveLength(1);
    expect(inputs[0].name).toBe("p1");
  });
  it("should return outputs", async () => {
    const net = makeNet();
    const outputs = controller.outputs(net, net.transitions[0]);
    expect(outputs).toHaveLength(1);
    expect(outputs[0].name).toBe("p2");
  });
  it("should return marking", async () => {
    const net = makeNet();
    const marking = controller.marking(net);
    for (const place of net.places) {
      expect(marking.get(place.id)).toBe(place.marking);
    }
  });
});