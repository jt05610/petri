import type { Place, Transition, Arc } from "@prisma/client";
import type {
  EventDetails,
  EventDetailsWithEnabled,
  NetDetailsWithChildren,
  TransitionWithEvents
} from "~/models/net.server";
import type { ColorProfile } from "~/lib/components/markedNet";
import cloneDeep from "lodash/cloneDeep";

export type Marking = {
  [placeID: string]: number;
}

type Node = Pick<Place, "id" | "name" | "bound"> | Pick<Transition, "id" | "name">;

type GraphVizParams = {
  rankdir?: "LR" | "TB" | "BT" | "RL";
}

type DeviceWithEvents = {
  id: string;
  name: string;
  instances: {
    id: string;
    name: string;
    addr: string;
  }[]
  events: {
    id: string;
    name: string;
    enabled?: boolean;
    fields: {
      id: string;
      name: string;
      type: "string" | "number" | "boolean" | string
    }[]
  }[]
}

export class PetriNet {
  net: NetDetailsWithChildren;
  events: EventDetails[];
  deviceIDFromInstanceID: {
    [instanceID: string]: string
  };
  devices: DeviceWithEvents[];
  initialMarking: Marking;

  constructor(net: NetDetailsWithChildren) {
    this.net = net;
    this.events = this.net.transitions.flatMap(transition => {
      if (!transition.events || transition.events.length === 0) {
        return [];
      }
      return transition.events.map(event => ({
        ...event,
        enabled: false
      }));
    });
    this.devices = [];

    this.deviceIDFromInstanceID = {};
    if (this.net.devices && this.net.devices.length > 0) {
      this.net.devices.forEach((device => {
        this.devices.push({
          id: device.id,
          name: device.name,
          instances: device.instances || [],
          events: this.events
        });
        if (!device.instances) {
          return;
        }
        device.instances.forEach(instance => {
          if (!this.net.devices || !instance.id) {
            return;
          }
          this.deviceIDFromInstanceID[instance.id] = device.id;
        });
      }));

    }
    if (this.net.children) {
      this.net.children.forEach(child => {
        if (child.devices && child.devices.length > 0) {
          child.devices.forEach(device => {
            this.devices.push({
              id: device.id,
              name: device.name,
              instances: device.instances || [],
              events: this.events
            });
            if (!device.instances) {
              return;
            }
            device.instances.forEach(instance => {
              if (!child.devices || !instance.id) {
                return;
              }
              this.deviceIDFromInstanceID[instance.id] = device.id;
            });
          });
        }
      });
    }

    this.initialMarking = {};
    this.net.places.forEach((place, i) => {
      this.initialMarking[place.id] = this.net.initialMarking[i];
    });
  }

  deviceIndexFromID(deviceID: string) {
    return this.devices.findIndex(device => device.id === deviceID);
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

  instanceOf(instanceID: string): string {
    return this.deviceIDFromInstanceID[instanceID];
  }

  allEvents(marking: Marking): EventDetailsWithEnabled[] {
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

  childDeviceEvents(marking: Marking): DeviceWithEvents[] {
    let ret: DeviceWithEvents[] = [];
    for (let child of this.net.children) {
      if (child.devices) {
        child.devices.forEach(device => {
          if (!device.instances) {
            return;
          }
          ret.push({
            id: device.id,
            name: device.name,
            instances: device.instances,
            events: child.transitions.flatMap(({ events }) => events!.map(event => ({
              ...event,
              enabled: this.enabledTransitions(marking).some(transition => transition.events && transition.events.some(e => e.id === event.id))
            })))
          });
        });

      }
    }
    return ret;
  }

  eventEnabled(marking: Marking, eventID: string) {
    return this.enabledTransitions(marking).some(transition => transition.events && transition.events.some(event => event.id === eventID));
  }

  fireEvent(marking: Marking, eventID: string) {
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
  }

  // handle the event and keep calling hot transitions until there are no more hot transitions
  handleEvent(marking: Marking, eventID: string): Marking {
    let newMarking = cloneDeep(marking);
    if (!this.eventEnabled(newMarking, eventID)) {
      console.log("Event: " + eventID + " is not enabled");
      return newMarking;
    }
    this.fireEvent(newMarking, eventID);
    while (this.hotTransitions(newMarking).length > 0) {
      this.fire(newMarking, this.hotTransitions(newMarking)[0]);
    }
    return newMarking;
  }

  // return a graphviz dot string representing the petri net. places are represented by circles, and transitions are represented by rectangles.
  toGraphViz(placeSize: number, colorProfile: ColorProfile, params: GraphVizParams = {}) {
    // make a marking where each place has a 0 marking
    const marking: Marking = {};
    this.net.places.forEach(place => {
      marking[place.id] = 0;
    });
    // then return the graphviz string with the 0 marking
    return this.toGraphVizWithMarking(placeSize, colorProfile, marking, params);
  }

  makePlaceNode(colorProfile: ColorProfile, place: Pick<Place, "id" | "name">, marking: Marking): string {
    return `    "p${place.id}" [color="${colorProfile.border}" label="${place.name}"  style=filled fillcolor="${marking[place.id] > 0 ? colorProfile.placeTokenColor : colorProfile.placeColor}"];`;
  }

  makeTransitionNode(colorProfile: ColorProfile, transition: Pick<Transition, "id" | "name">, marking: Marking): string {
    return `    "t${transition.id}" [color="${colorProfile.border}" label="${transition.name}" shape=box style=filled fillcolor="${this.enabledTransitions(marking).includes(transition) ? colorProfile.enabledTransition : colorProfile.transitionColor}"];`;
  }

  makeArc(colorProfile: ColorProfile, arc: Pick<Arc, "fromPlace" | "placeID" | "transitionID">, marking: Marking): string {
    const baseString = arc.fromPlace ? `    "p${arc.placeID}" -> "t${arc.transitionID}"` : `    "t${arc.transitionID}" -> "p${arc.placeID}"`;
    if (this.enabledTransitions(marking).some(transition => transition.id === arc.transitionID)) {
      return `${baseString} [color="${colorProfile.enabledArc}"];`;
    }
    return `${baseString} [color="${colorProfile.border}"];`;
  }

  // return a graphviz dot string representing the petri net with the passed marking. tokens in the markings are represented by black dots.
  toGraphVizWithMarking(placeSize: number, colorProfile: ColorProfile, marking: Marking, params: GraphVizParams = {}) {
    const { rankdir = "LR" } = params;
    const graph = [
      "digraph {",
      `    bgcolor="${colorProfile.bg}"`,
      `    rankdir=${rankdir};`,
      `    node [shape=circle fontname="Arial" fontsize=10 fixedsize=false fontcolor="${colorProfile.text}" width=${placeSize}];`,
      ...this.net.places.map(place => this.makePlaceNode(colorProfile, place, marking)),
      ...this.net.transitions.map(transition => this.makeTransitionNode(colorProfile, transition, marking)),
      ...this.net.arcs.map(arc => this.makeArc(colorProfile, arc, marking)),
      "}"
    ];
    return graph.join("\n");
  }

  // combines all child nets into a single net. For the placeInterfaces and transitionInterfaces, the id and name are the id and name of the combined node in the full net, and the corresponding nodes in the child nets are stored in the transitions or places array.
  get combinedNet() {
    // all child places, transitions, or arcs except the ones that appear in the interfaces
    let places = [...this.net.places, ...this.net.children.map(child => child.places.filter((place) => !this.net.placeInterfaces.some(placeInterface => placeInterface.id === place.id))).flat()];
    let transitions = [...this.net.transitions, ...this.net.children.map(child => child.transitions).flat()];
    let arcs = [...this.net.arcs, ...this.net.children.map(child => child.arcs).flat()];
    const markingMap: {
      [placeID: string]: number
    } = {};
    this.net.initialMarking.forEach((marking, i) => {
      markingMap[this.net.places[i].id] = marking;
    });
    this.net.children.forEach(child => {
      child.initialMarking.forEach((marking, i) => {
        markingMap[child.places[i].id] = marking;
      });
    });
    this.net.placeInterfaces.forEach(placeInterface => {
      places.push({
        id: placeInterface.id,
        name: placeInterface.name,
        bound: placeInterface.bound
      });
      // iterate through the interface's sub places and for any arc that is connected to this place, replace that arcs place ID with this arc
      placeInterface.places.forEach(place => {
        arcs.forEach(arc => {
          if (arc.placeID === place.id) {
            arc.placeID = placeInterface.id;
          }
        });
        // remove the place
      });
      // filter all places where the id that appear in placesInterfaces.places from the main places array
      places = places.filter(place => !placeInterface.places.some(placeInterfacePlace => placeInterfacePlace.id === place.id));
    });
    this.net.transitionInterfaces.forEach(transitionInterface => {
      transitions.push({
        id: transitionInterface.id,
        name: transitionInterface.name,
        events: transitionInterface.events
      });
      // iterate through the interface's sub transitions and for any arc that is connected to this transition, replace that arcs transition ID with this arc
      transitionInterface.transitions.forEach(transition => {
        arcs.forEach(arc => {
          if (arc.transitionID === transition.id) {
            arc.transitionID = transitionInterface.id;
          }
        });
      });
      // filter all transitions where the id that appear in transitionInterfaces.transitions from the main transitions array
      transitions = transitions.filter(transition => !transitionInterface.transitions.some(transitionInterfaceTransition => transitionInterfaceTransition.id === transition.id));
    });

    const initialMark = places.map(place => {
      return markingMap[place.id] || 0;
    });
    return new PetriNet({
      description: this.net.description,
      places, transitions, arcs,
      initialMarking: initialMark,
      devices: this.net.devices,
      placeInterfaces: [],
      transitionInterfaces: [],
      children: this.net.children,
      name: this.net.name,
      id: this.net.id
    });
  }
}
