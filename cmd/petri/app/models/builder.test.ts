import { expect } from "vitest";
import { builder } from "~/models/builder";
import { isCuid } from "@paralleldrive/cuid2";

const fakePlace = {
  name: "p1",
  bound: 1,
  initialMarking: 0,
  marking: 0
};

describe.concurrent("Builder", async () => {
  describe.concurrent("places", async () => {
    it("should create a new place", async () => {
      let net = builder.newNet("test");
      net = builder.addPlace(net, fakePlace);
      const [place] = net.places;
      expect(place.name).toBe(fakePlace.name);
      expect(place.bound).toBe(fakePlace.bound);
      expect(place.initialMarking).toBe(fakePlace.initialMarking);
      expect(isCuid(place.id)).toBe(true);
    });
    it("should remove a place", async () => {
      let net = builder.newNet("test");
      net = builder.addPlace(net, fakePlace);
      const [place] = net.places;
      if (!place) {
        throw new Error("Place not found");
      }
      net = builder.removePlace(net, place);
      expect(net.places).toHaveLength(0);
    });
    it("should update a place", async () => {
      let net = builder.newNet("test");
      net = builder.addPlace(net, fakePlace);
      const [place] = net.places;
      if (!place) {
        throw new Error("Place not found");
      }
      net = builder.updatePlace(net, {
        ...place,
        name: "p2"
      });
      expect(net.places[0].name).toBe("p2");
    });
  });
  describe.concurrent("transitions", async () => {
    it("should create a new transition", async () => {
      let net = builder.newNet("test");
      net = builder.addTransition(net, {
        name: "t1",
        events: []
      });
      const [transition] = net.transitions;
      expect(transition.name).toBe("t1");
      expect(isCuid(transition.id)).toBe(true);
    });
    it("should remove a transition", async () => {
      let net = builder.newNet("test");
      net = builder.addTransition(net, {
        name: "t1",
        events: []
      });
      const [transition] = net.transitions;
      if (!transition) {
        throw new Error("Transition not found");
      }
      net = builder.removeTransition(net, transition);
      expect(net.transitions).toHaveLength(0);
    });
    it("should update a transition", async () => {
      let net = builder.newNet("test");
      net = builder.addTransition(net, {
        name: "t1",
        events: []
      });
      const [transition] = net.transitions;
      if (!transition) {
        throw new Error("Transition not found");
      }
      net = builder.updateTransition(net, {
        ...transition,
        name: "t2"
      });
      expect(net.transitions[0].name).toBe("t2");
    });
  });
  describe.concurrent("arcs", async () => {
    it("should create a new arc", async () => {
      let net = builder.newNet("test");
      net = builder.addPlace(net, fakePlace);
      net = builder.addTransition(net, {
        name: "t1",
        events: []
      });
      net = builder.addArc(net, {
        from: net.places[0],
        to: net.transitions[0]
      });
      const [arc] = net.arcs;
      expect(arc.from).toBe(net.places[0]);
      expect(arc.to).toBe(net.transitions[0]);
      expect(isCuid(arc.id)).toBe(true);
    });
    it("should remove an arc", async () => {
      let net = builder.newNet("test");
      net = builder.addPlace(net, fakePlace);
      net = builder.addTransition(net, {
        name: "t1",
        events: []
      });
      net = builder.addArc(net, {
        from: net.places[0],
        to: net.transitions[0]
      });
      const [arc] = net.arcs;
      if (!arc) {
        throw new Error("Arc not found");
      }
      net = builder.removeArc(net, arc);
      expect(net.arcs).toHaveLength(0);
    });
    it("shouldn't create an arc between two places", async () => {
      let net = builder.newNet("test");
      net = builder.addPlace(net, fakePlace);
      net = builder.addPlace(net, {
        name: "p2",
        bound: 1,
        initialMarking: 0,
        marking: 0
      });
      expect(() => builder.addArc(net, {
          from: net.places[0],
          to: net.places[1]
        })
      ).toThrowError(/same type/);
    });
  });

  describe.concurrent("net", async () => {
    it("should create a new net", async () => {
      const net = builder.newNet("test");
      expect(net).toEqual({
        id: expect.any(String),
        name: "test",
        places: [],
        transitions: [],
        arcs: [],
        children: []
      });
    });
    it("should copy a net", async () => {
      const net = builder.newNet("test");
      const copy = builder.copy(net);
      expect(copy).toEqual({
        id: expect.any(String),
        name: "test (1)",
        places: [],
        transitions: [],
        arcs: [],
        children: []
      });
    });
    it("should add a child", async () => {
      const net = builder.newNet("test");
      const child = builder.newNet("child");
      const parent = builder.addChild(net, child);
      expect(parent.children).toHaveLength(1);
      expect(parent.children[0].name).toBe("child");
    });
    it("should remove a child", async () => {
      const net = builder.newNet("test");
      const child = builder.newNet("child");
      let parent = builder.addChild(net, child);
      parent = builder.removeChild(parent, child);
      expect(parent.children).toHaveLength(0);
    });
    it("should rename", async () => {
      const net = builder.newNet("test");
      const renamed = builder.rename(net, "test2");
      expect(renamed.name).toBe("test2");
    });
  });
});
