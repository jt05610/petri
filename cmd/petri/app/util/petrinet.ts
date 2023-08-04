import type { Place, Transition } from "@prisma/client";
import type { NetDetails, TransitionWithEvents } from "~/models/net.server";

type Marking = {
  [placeID: string]: number;
}

type Node = Pick<Place, "id" | "name" | "bound"> | Pick<Transition, "id" | "name">;

export class PetriNet {
  net: NetDetails;

  constructor(net: NetDetails) {
    this.net = net;
  }

  get nodes() {
    return [...this.net.places, ...this.net.transitions];
  }

  get arcs() {
    return this.net.arcs;
  }

  // get all inputs to the passed node.
  inputs(node: Node) {
    // if node is a place type then return all arcs where placeID is node.id and fromPlace is false
    if ("bound" in node) {
      return this.arcs.filter(arc => arc.placeID === node.id && !arc.fromPlace);
    }
    // if node is a transition type then return all arcs where transitionID is node.id and fromPlace is true
    return this.arcs.filter(arc => arc.transitionID === node.id && arc.fromPlace);
  }

  // get all outputs to the passed node.
  outputs(node: Node) {
    // if node is a place type then return all arcs where placeID is node.id and fromPlace is true
    if ("bound" in node) {
      return this.arcs.filter(arc => arc.placeID === node.id && arc.fromPlace);
    }
    // if node is a transition type then return all arcs where transitionID is node.id and fromPlace is false
    return this.arcs.filter(arc => arc.transitionID === node.id && !arc.fromPlace);
  }

  enabledTransitions(marking: Marking) {
    return this.net.transitions.filter(transition => this.inputs(transition).every(arc => marking[arc.placeID] > 0));
  }

  hotTransitions(marking: Marking) {
    // returns transitions that are enabled and do not have an associated event, thus, they fire immediately
    return this.enabledTransitions(marking).filter(transition => !transition.events || transition.events.length === 0);
  }

  fire(marking: Marking, transition: TransitionWithEvents): Marking {
    if (!transition) {
      throw new Error("Transition does not exist");
    }

    if (!this.enabledTransitions(marking).includes(transition)) {
      throw new Error("Transition is not enabled");
    }
    this.inputs(transition).forEach(arc => {
      marking[arc.placeID]--;
    });
    this.outputs(transition).forEach(arc => {
      marking[arc.placeID]++;
    });
    return marking;
  }

  allEvents(marking: Marking) {
    return this.net.transitions.flatMap(transition => {
      if (!transition.events || transition.events.length === 0) {
        return [];
      }
      return transition.events.map(event => ({
        ...event,
        enabled: this.enabledTransitions(marking).includes(transition)
      }));
    });
  }

  eventEnabled(marking: Marking, eventID: string) {
    return this.enabledTransitions(marking).some(transition => transition.events && transition.events.some(event => event.id === eventID));
  }

  fireEvent(marking: Marking, eventID: string): Marking {
    const transition = this.net.transitions.find(transition => transition.events && transition.events.some(event => event.id === eventID));
    if (!transition) {
      throw new Error("Transition does not exist");
    }

    if (!this.enabledTransitions(marking).includes(transition)) {
      throw new Error("Transition is not enabled");
    }
    this.inputs(transition).forEach(arc => {
      marking[arc.placeID]--;
    });
    this.outputs(transition).forEach(arc => {
      marking[arc.placeID]++;
    });
    return marking;
  }

  // handle the event and keep calling hot transitions until there are no more hot transitions
  handleEvent(marking: Marking, eventID: string): Marking {
    console.log("Handling event: " + eventID + " with marking: " + JSON.stringify(marking));
    if (!this.eventEnabled(marking, eventID)) {
      console.log("Event: " + eventID + " is not enabled")
      return marking
    }
    marking = this.fireEvent(marking, eventID);
    while (this.hotTransitions(marking).length > 0) {
      marking = this.fire(marking, this.hotTransitions(marking)[0]);
    }
    console.log("Handled event: " + eventID + " with marking: " + JSON.stringify(marking));
    return marking;
  }
}
