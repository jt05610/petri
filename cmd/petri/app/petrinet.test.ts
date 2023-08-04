import { PetriNet } from "~/util/petrinet";

test("PetriNet", () => {
  const net = new PetriNet({
    id: "1",
    name: "test",
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
  expect(net).toBeInstanceOf(PetriNet);

  expect(net.nodes).toHaveLength(4);
  expect(net.arcs).toHaveLength(4);
  expect(net.enabledTransitions).toHaveLength(1);

  expect(net.eventEnabled("1")).toBeTruthy();
  expect(net.enabledTransitions[0].id).toBe("2");
  // expect firing transition 1 to not fire
  expect(() => net.fire(net.net.transitions[0])).toThrow();
  // expect firing transition 2 to fire
  expect(() => net.fire(net.net.transitions[1])).not.toThrow();
  expect(net.eventEnabled("1")).toBeFalsy();
  // now we expect transition 2 to be disabled and transition 1 to be enabled
  expect(() => net.fire(net.net.transitions[1])).toThrow();
  expect(net.enabledTransitions).toHaveLength(1);
  expect(net.enabledTransitions[0].id).toBe("1");
  expect(() => net.fire(net.net.transitions[0])).not.toThrow();
  // when we handle the event for transition 2, we expect it to be enabled again as transition 1 is hot and will fire immediately
  expect(() => net.handleEvent("1")).not.toThrow();
  expect(() => net.handleEvent("1")).not.toThrow();
  expect(() => net.handleEvent("1")).not.toThrow();

});